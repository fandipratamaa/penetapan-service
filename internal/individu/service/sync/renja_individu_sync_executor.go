package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bappeda-dev-team/penetapan-service/internal/common"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/client"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/domain"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/repository"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/web"
	"github.com/bappeda-dev-team/penetapan-service/internal/kode"
)

type RenjaIndividuSyncExecutor struct {
	Repo   *repository.PenetapanIndividuRepository
	Client *client.Client
	Logger *slog.Logger
}

func NewRenjaIndividuSyncExecutor(
	repo *repository.PenetapanIndividuRepository,
	client *client.Client,
	logger *slog.Logger,
) *RenjaIndividuSyncExecutor {
	return &RenjaIndividuSyncExecutor{
		Repo:   repo,
		Client: client,
		Logger: logger,
	}
}

func (ex *RenjaIndividuSyncExecutor) Sync(
	ctx context.Context,
	syncId int64,
	req *web.SyncPenetapanRequest,
	currentUser string,
) (web.SyncPenetapanSummary, error) {
	request := client.SyncRequest{
		PegawaiId: req.PegawaiId,
		KodeOpd:   req.KodeOpd,
		Tahun:     req.Tahun,
	}
	// outsource
	renjaIndividus, err := ex.Client.SyncRenjaIndividu(ctx, request)
	if err != nil {
		return web.SyncPenetapanSummary{}, err
	}

	if !hasValidRenjaIndividu(renjaIndividus) {
		return web.SyncPenetapanSummary{}, common.NewValidation(
			"tidak ada data penetapan renja individu yang siap disinkronkan",
		)
	}

	// db transaction
	tx, err := ex.Repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.SyncPenetapanSummary{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	snapshot := domain.SnapshotPenetapan{
		JenisSnapshot: domain.JenisPenetapanRenjaIndividu,
		PegawaiId:     req.PegawaiId,
		KodeOpd:       req.KodeOpd,
		Tahun:         req.Tahun,
	}

	snapshotId, err := ex.createActiveSnapshot(ctx, tx, snapshot, currentUser)
	if err != nil {
		return web.SyncPenetapanSummary{}, err
	}

	kodeSubkegiatans := make([]string, 0, len(renjaIndividus))
	for _, renja := range renjaIndividus {
		if strings.TrimSpace(renja.KodeSubkegiatan) != "" {
			kodeSubkegiatans = append(kodeSubkegiatans, renja.KodeSubkegiatan)
		}
	}

	renjaOpdRefs, err := ex.Repo.ResolveRenjaOpdIds(
		ctx, tx, snapshot.KodeOpd, snapshot.Tahun, kodeSubkegiatans,
	)
	if err != nil {
		return web.SyncPenetapanSummary{}, err
	}

	// SAVE RENJA INDIVIDU
	summary := SummaryCounter{}
	for _, renja := range renjaIndividus {
		kodePaguProgram := fmt.Sprintf("PAGU-PRG-%s-%d-%s", renja.KodeProgram, snapshot.Tahun, "programs")
		kodePaguKegiatan := fmt.Sprintf("PAGU-KEG-%s-%d-%s", renja.KodeKegiatan, snapshot.Tahun, "kegiatans")
		kodePaguSubkegiatan := fmt.Sprintf("PAGU-SUBKEG-%s-%d-%s", renja.KodeSubkegiatan, snapshot.Tahun, "subkegiatans")

		ref := renjaOpdRefs[renja.KodeSubkegiatan]
		renjaIndividu := domain.RenjaIndividu{
			PegawaiId:  snapshot.PegawaiId,
			KodeOpd:    snapshot.KodeOpd,
			TahunAktif: snapshot.Tahun,

			KodePk:        renja.IdRekinPemilikPk,
			NamaPemilikPk: renja.NamaPemilikPk,
			LevelPk:       renja.LevelPk,

			KodeProgram:     renja.KodeProgram,
			NamaProgram:     renja.NamaProgram,
			KodePaguProgram: kodePaguProgram,
			PaguProgram:     renja.PaguProgram,

			KodeKegiatan:     renja.KodeKegiatan,
			NamaKegiatan:     renja.NamaKegiatan,
			KodePaguKegiatan: kodePaguKegiatan,
			PaguKegiatan:     renja.PaguKegiatan,

			KodeSubkegiatan:     renja.KodeSubkegiatan,
			NamaSubkegiatan:     renja.NamaSubkegiatan,
			KodePaguSubkegiatan: kodePaguSubkegiatan,
			PaguSubkegiatan:     renja.PaguSubkegiatan,

			PenetapanIndividuId: snapshotId,
			CreatedBy:           &currentUser,

			KodeRenjaOpdProgram:     kode.KodeRenjaOpd(ref.IdProgram),
			KodeRenjaOpdKegiatan:    kode.KodeRenjaOpd(ref.IdKegiatan),
			KodeRenjaOpdSubkegiatan: kode.KodeRenjaOpd(ref.IdSubkegiatan),
		}
		idRenja, err := ex.Repo.SaveRenjaIndividu(ctx, tx, renjaIndividu)
		if err != nil {
			return web.SyncPenetapanSummary{}, err
		}
		summary.AddRenjaIndividu(1)
		// INDIKATOR
		if err := ex.saveIndikatorRenja(
			ctx,
			tx,
			idRenja,
			repository.TierRenjaProgram,
			ref.IdProgram,
			snapshot.Tahun,
			renja.IndikatorPrograms,
			&summary,
		); err != nil {
			return web.SyncPenetapanSummary{}, err
		}

		if err := ex.saveIndikatorRenja(
			ctx,
			tx,
			idRenja,
			repository.TierRenjaKegiatan,
			ref.IdKegiatan,
			snapshot.Tahun,
			renja.IndikatorKegiatans,
			&summary,
		); err != nil {
			return web.SyncPenetapanSummary{}, err
		}

		if err := ex.saveIndikatorRenja(
			ctx,
			tx,
			idRenja,
			repository.TierRenjaSubkegiatan,
			ref.IdSubkegiatan,
			snapshot.Tahun,
			renja.IndikatorSubkegiatans,
			&summary,
		); err != nil {
			return web.SyncPenetapanSummary{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return web.SyncPenetapanSummary{}, err
	}

	return summary.Response(), nil
}

func (ex *RenjaIndividuSyncExecutor) createActiveSnapshot(
	ctx context.Context,
	tx *sql.Tx,
	snapshot domain.SnapshotPenetapan,
	currentUser string,
) (int64, error) {
	versi, err := ex.Repo.GetPenetapanNextVersion(ctx, tx, snapshot)
	if err != nil {
		return 0, err
	}

	if versi > 1 {
		if err := ex.Repo.DeactivateOldSnapshot(ctx, tx, snapshot); err != nil {
			return 0, err
		}
	}

	snapshot.Versi = versi
	snapshot.SnapshotStatus = domain.SnapshotStatusActive
	snapshot.GeneratedBy = &currentUser
	snapshot.IsActive = true

	snapshotID, err := ex.Repo.SaveSnapshot(ctx, tx, snapshot)
	if err != nil {
		ex.Logger.Error("save renja individu error", "snapshot", snapshot)
		return 0, err
	}

	return snapshotID, nil
}

func (ex *RenjaIndividuSyncExecutor) saveIndikatorRenja(
	ctx context.Context,
	tx *sql.Tx,
	renjaID int64,
	jenis string,
	parentOpdID int64,
	tahun int,
	indikators []client.IndikatorRenja,
	summary *SummaryCounter,
) error {
	indikatorOpdRef, err := ex.Repo.ResolveIndikatorRenjaOpd(
		ctx, tx, jenis, parentOpdID, tahun,
	)
	if err != nil {
		return err
	}

	for _, ind := range indikators {
		kodeIndikator := fmt.Sprintf("IND-%s", ind.Id)

		// Kode didahulukan atas nama: nama indikator di api_ekak ("INDIKATOR
		// PROGRAM A") dan di cascade renja OPD sama sekali berbeda, padahal
		// kode_indikator-nya identik setelah prefix IND- dibuang.
		indikatorOpdID, _ := indikatorOpdRef.Lookup(ind.Id, ind.Indikator)

		kodeIndikatorOpd := kode.KodeIndikatorRenjaOpd(indikatorOpdID)

		indikator := domain.IndikatorRenjaIndividu{
			RenjaIndividuID: renjaID,
			JenisIndikator:  jenis,
			// WARNING KODE INDIKATOR
			KodeIndikatorRenja:    kodeIndikator,
			Indikator:             ind.Indikator,
			KodeIndikatorRenjaOpd: kodeIndikatorOpd,
		}

		idIndikator, err := ex.Repo.SaveIndikatorRenjaIndividu(ctx, tx, indikator)
		if err != nil {
			return err
		}

		summary.AddIndikatorRenjaIndividu(1)

		targetOpdIDs := map[int]int64{}

		if indikatorOpdID > 0 {
			targetOpdIDs, err = ex.Repo.ResolveTargetRenjaOpd(
				ctx, tx, jenis, indikatorOpdID,
			)
			if err != nil {
				return err
			}
		}

		for _, target := range ind.Targets {
			kodeTarget := fmt.Sprintf("TGT-%s", target.Id)
			targetFloat, err := ParseTargetFloat(target.Target)
			if err != nil {
				return fmt.Errorf(
					"invalid target indikator %q value %q",
					ind.Indikator,
					target.Target,
				)
			}

			var kodeTargetOpd string
			if opdID, ok := targetOpdIDs[target.Tahun]; ok {
				kodeTargetOpd = kode.KodeTargetRenjaOpd(opdID)
			}

			targetRenja := domain.TargetRenjaIndividu{
				IndikatorRenjaIndividuID: idIndikator,
				JenisTarget:              jenis,
				KodeTargetRenja:          kodeTarget,
				KodeTargetRenjaOpd:       kodeTargetOpd,
				Target:                   targetFloat,
				Satuan:                   target.Satuan,
				Tahun:                    target.Tahun,
			}

			_, err = ex.Repo.SaveTargetRenjaIndividu(ctx, tx, targetRenja)
			if err != nil {
				return err
			}

			summary.AddTargetRenjaIndividu(1)
		}
	}

	return nil
}

func ParseTargetFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)

	// Hanya ubah 10,5 -> 10.5
	if strings.Count(s, ",") == 1 &&
		!strings.Contains(s, ".") {
		s = strings.Replace(s, ",", ".", 1)
	}

	return strconv.ParseFloat(s, 64)
}

func hasValidRenjaIndividu(renjaIndividus []client.RenjaIndividuResponse) bool {
	for _, renja := range renjaIndividus {
		if hasRenjaIndividuContent(renja.NamaProgram, renja.PaguProgram, renja.IndikatorPrograms) ||
			hasRenjaIndividuContent(renja.NamaKegiatan, renja.PaguKegiatan, renja.IndikatorKegiatans) ||
			hasRenjaIndividuContent(renja.NamaSubkegiatan, renja.PaguSubkegiatan, renja.IndikatorSubkegiatans) {
			return true
		}
	}
	return false
}

func hasRenjaIndividuContent(nama string, pagu int64, indikators []client.IndikatorRenja) bool {
	if strings.TrimSpace(nama) == "" {
		return false
	}
	if pagu > 0 {
		return true
	}
	for _, ind := range indikators {
		if strings.TrimSpace(ind.Indikator) == "" || len(ind.Targets) == 0 {
			continue
		}
		for _, tgt := range ind.Targets {
			target, err := ParseTargetFloat(tgt.Target)
			if err != nil {
				continue
			}
			if target != 0 && strings.TrimSpace(tgt.Satuan) != "" {
				return true
			}
		}
	}
	return false
}
