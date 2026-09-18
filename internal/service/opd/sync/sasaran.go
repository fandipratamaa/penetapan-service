package sync

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bappeda-dev-team/penetapan-service/internal/client/perencanaan"
	"github.com/bappeda-dev-team/penetapan-service/internal/common"
	"github.com/bappeda-dev-team/penetapan-service/internal/kode"
	"github.com/bappeda-dev-team/penetapan-service/internal/model/domain"
	"github.com/bappeda-dev-team/penetapan-service/internal/model/web"
	"github.com/bappeda-dev-team/penetapan-service/internal/repository"
	"github.com/bappeda-dev-team/penetapan-service/internal/service/opd/helper"
)

type SasaranSyncExecutor struct {
	Repo              *repository.PenetapanOpdRepository
	PerencanaanClient *perencanaan.PerencanaanClient
	Logger            *slog.Logger
}

func NewSasaranSyncExecutor(
	repo *repository.PenetapanOpdRepository,
	perencanaanClient *perencanaan.PerencanaanClient,
	logger *slog.Logger,
) *SasaranSyncExecutor {
	return &SasaranSyncExecutor{
		Repo:              repo,
		PerencanaanClient: perencanaanClient,
		Logger:            logger,
	}
}

func (ex *SasaranSyncExecutor) Sync(
	ctx context.Context,
	syncId int64,
	req *web.SyncPenetapanOpdRequest,
	currentUser string,

) (web.SyncPenetapanOpdSummary, error) {
	perencanaanRequest := perencanaan.PerencanaanRequest{
		KodeOpd: req.KodeOpd,
		Tahun:   req.Tahun,
	}

	sasaranPerencanaans, err := ex.PerencanaanClient.GetPenetapanSasaranOpd(ctx, perencanaanRequest)
	if err != nil {
		return web.SyncPenetapanOpdSummary{}, err
	}

	if !hasValidSasaranOpd(sasaranPerencanaans) {
		return web.SyncPenetapanOpdSummary{}, common.NewValidation(
			"tidak ada data penetapan sasaran OPD yang siap disinkronkan",
		)
	}

	// mulai butuh tx
	tx, err := ex.Repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.SyncPenetapanOpdSummary{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	versiPenetapan, err := ex.Repo.GetPenetapanNextVersion(
		ctx, tx, req.KodeOpd, domain.JenisPenetapanSasaran, req.Tahun)
	if err != nil {
		return web.SyncPenetapanOpdSummary{}, err
	}
	if versiPenetapan > 1 {
		errDeact := ex.Repo.DeactivateOldSnapshot(
			ctx, tx, req.KodeOpd, domain.JenisPenetapanSasaran, req.Tahun)
		if errDeact != nil {
			return web.SyncPenetapanOpdSummary{}, err
		}
	}

	// simpan snapshot penetapan
	snapshot := domain.PenetapanOpd{
		KodeOpd:        req.KodeOpd,
		Tahun:          req.Tahun,
		JenisPenetapan: domain.JenisPenetapanSasaran,
		Versi:          versiPenetapan,
		SnapshotStatus: domain.SnapshotStatusActive,
		GeneratedBy:    &currentUser,
		IsActive:       true,
	}

	penetapanId, err := ex.Repo.SavePenetapanOpd(ctx, tx, snapshot)
	if err != nil {
		ex.Logger.Error("save penetapan opd error",
			"snapshot", snapshot)
		return web.SyncPenetapanOpdSummary{}, err
	}

	// convert response perencanaan to snapshot
	snapshotSasaran, err := ex.toSasaranSnapshots(sasaranPerencanaans, currentUser, penetapanId, req.Tahun)
	if err != nil {
		return web.SyncPenetapanOpdSummary{}, err
	}

	// simpan snapshot
	var jumlahSasaran int
	var jumlahIndikator int
	var jumlahTarget int
	for _, sasaran := range snapshotSasaran {
		sasaranId, err := ex.Repo.SaveSasaranPenetapanOpd(ctx, tx, sasaran)
		if err != nil {
			return web.SyncPenetapanOpdSummary{}, err
		}
		jumlahSasaran += 1

		// indikator
		for _, indikator := range sasaran.Indikator {

			indId, err := ex.Repo.SaveIndikatorSasaranPenetapanOpd(ctx, tx, indikator, sasaranId)
			if err != nil {
				return web.SyncPenetapanOpdSummary{}, err
			}
			jumlahIndikator += 1

			// target
			targets := indikator.Target
			tgtInserted, err := ex.Repo.SaveTargetIndikatorSasaranBatch(ctx, tx, indId, targets)
			if err != nil {
				return web.SyncPenetapanOpdSummary{}, err
			}
			jumlahTarget += tgtInserted
		}
	}

	// commit transaction penetapan tujuan
	err = tx.Commit()
	if err != nil {
		return web.SyncPenetapanOpdSummary{}, err
	}

	return web.SyncPenetapanOpdSummary{
		Sasaran:   &jumlahSasaran,
		Indikator: jumlahIndikator,
		Target:    jumlahTarget,
	}, nil
}

func (ex *SasaranSyncExecutor) toSasaranSnapshots(sasaranPerencanaans []perencanaan.PerencanaanSasaranOpdResponse, currentUser string, penetapanId int64, tahunAktif int) ([]domain.SasaranPenetapanOpd, error) {

	createdBy := &currentUser
	var penetapanSasaranOpds = []domain.SasaranPenetapanOpd{}
	for _, per := range sasaranPerencanaans {
		for _, sasaran := range per.SasaranOpd {

			indikators := make(
				[]domain.IndikatorSasaranPenetapanOpd,
				0,
				len(sasaran.Indikator),
			)

			for _, ind := range sasaran.Indikator {
				indSnapshot, err := ex.toIndikatorSasaranSnapshot(
					ind, per.KodeOpd,
					createdBy,
					tahunAktif,
				)
				if err != nil {
					return nil, err
				}
				indikators = append(
					indikators,
					indSnapshot,
				)

			}
			sasSnapshot := ex.toSasaranSnapshot(
				sasaran,
				per.KodeOpd,
				tahunAktif,
				penetapanId,
				createdBy,
				indikators,
			)
			penetapanSasaranOpds = append(penetapanSasaranOpds,
				sasSnapshot,
			)
		}
	}

	return penetapanSasaranOpds, nil
}

func (ex *SasaranSyncExecutor) toSasaranSnapshot(
	sasaran perencanaan.SasaranOpdDetailResponse,
	kodeOpd string,
	tahunAktif int,
	penetapanId int64,
	createdBy *string,
	indikators []domain.IndikatorSasaranPenetapanOpd,
) domain.SasaranPenetapanOpd {
	kodeSasaranOpd := kode.KodeSasaranOpd(sasaran.Id)
	periodeTujuan := fmt.Sprintf("%s-%s",
		sasaran.TahunAwal,
		sasaran.TahunAkhir)
	kodeTujuan := fmt.Sprintf("TUJ-OPD-%d", sasaran.IdTujuanOpd)
	return domain.SasaranPenetapanOpd{
		KodeOpd:        kodeOpd,
		KodeTujuanOpd:  kodeTujuan,
		KodeSasaranOpd: kodeSasaranOpd,
		SasaranOpd:     sasaran.NamaSasaranOpd,
		Periode:        periodeTujuan,
		TahunAktif:     tahunAktif,
		CreatedBy:      createdBy,
		PenetapanId:    penetapanId,
		Indikator:      indikators,
	}
}

func (ex *SasaranSyncExecutor) toIndikatorSasaranSnapshot(ind perencanaan.IndikatorSasaranResponse, kodeOpd string, createdBy *string, tahunAktif int) (domain.IndikatorSasaranPenetapanOpd, error) {
	targets, err := ex.toTargetSnapshots(ind.Target, ind.NamaIndikator, createdBy)
	if err != nil {
		return domain.IndikatorSasaranPenetapanOpd{}, err
	}
	kodeIndikator := kode.KodeIndikatorSasaranOpd(ind.Id)
	return domain.IndikatorSasaranPenetapanOpd{
		KodeIndikator:       kodeIndikator,
		KodeOpd:             kodeOpd,
		Indikator:           ind.NamaIndikator,
		RumusPerhitungan:    &ind.RumusPerhitungan,
		SumberData:          &ind.SumberData,
		DefinisiOperasional: &ind.DefinisiOperasional,
		TahunAktif:          tahunAktif,
		CreatedBy:           createdBy,
		Target:              targets,
	}, nil
}

func (ex *SasaranSyncExecutor) toTargetSnapshots(targets []perencanaan.TargetResponse, namaIndikator string, createdBy *string) ([]domain.TargetIndikatorSasaranPenetapanOpd, error) {
	result := make([]domain.TargetIndikatorSasaranPenetapanOpd, 0, len(targets))
	for _, tgt := range targets {
		tahunTarget, errTahun := helper.ParseTahun(tgt.Tahun)
		if errTahun != nil {
			return nil, errTahun
		}
		target, errConv := helper.ParseFloat(tgt.TargetIndikator)
		if errConv != nil {
			return nil, errConv
		}
		tgtSnapshot := ex.toTargetIndikatorSasaranSnapshot(tgt, tahunTarget, target, createdBy)
		result = append(result, tgtSnapshot)
	}
	return result, nil
}

func (ex *SasaranSyncExecutor) toTargetIndikatorSasaranSnapshot(tgt perencanaan.TargetResponse, tahunTarget int, target float64, createdBy *string) domain.TargetIndikatorSasaranPenetapanOpd {
	kodeTarget := kode.KodeTargetSasaranOpd(tgt.Id)
	return domain.TargetIndikatorSasaranPenetapanOpd{
		KodeTarget: kodeTarget,
		Tahun:      tahunTarget,
		Target:     target,
		Satuan:     tgt.SatuanIndikator,
		CreatedBy:  createdBy,
	}
}

func hasValidSasaranOpd(sasaranPerencanaans []perencanaan.PerencanaanSasaranOpdResponse) bool {
	for _, per := range sasaranPerencanaans {
		for _, sasaran := range per.SasaranOpd {
			if strings.TrimSpace(sasaran.NamaSasaranOpd) == "" {
				continue
			}
			for _, indikator := range sasaran.Indikator {
				if strings.TrimSpace(indikator.NamaIndikator) == "" || len(indikator.Target) == 0 {
					continue
				}
				for _, target := range indikator.Target {
					if isValidTargetOpd(target) {
						return true
					}
				}
			}
		}
	}
	return false
}
