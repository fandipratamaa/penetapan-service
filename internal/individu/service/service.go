package service

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/bappeda-dev-team/penetapan-service/internal/individu/client"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/domain"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/repository"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/service/sync"
	"github.com/bappeda-dev-team/penetapan-service/internal/individu/web"
)

type PenetapanIndividuService struct {
	Repo         *repository.PenetapanIndividuRepository
	Client       *client.Client
	SyncExecutor *sync.Registry
	Logger       *slog.Logger
}

func NewPenetapanIndividuService(
	repo *repository.PenetapanIndividuRepository,
	client *client.Client,
	syncExecutor *sync.Registry,
	// logger taruh paling bawah
	logger *slog.Logger,
) *PenetapanIndividuService {
	return &PenetapanIndividuService{
		Repo:         repo,
		Client:       client,
		SyncExecutor: syncExecutor,
		Logger:       logger,
	}
}

func (s *PenetapanIndividuService) FindRekinsIndividu(
	ctx context.Context,
	req web.SyncPenetapanRequest,
) (web.RekinPenetapanIndividuResponse, error) {
	s.Logger.Info("FindRekinsIndividu")

	snapshot := domain.SnapshotPenetapan{
		JenisSnapshot:  domain.JenisPenetapanPk,
		PegawaiId:      req.PegawaiId,
		KodeOpd:        req.KodeOpd,
		Tahun:          req.Tahun,
		SnapshotStatus: domain.SnapshotStatusActive,
	}
	activeSnapshot, err := s.getActiveSnapshot(ctx, snapshot)
	if err != nil {
		return web.RekinPenetapanIndividuResponse{}, err
	}
	snapshot.SnapshotId = &activeSnapshot.Id
	snapshot.Versi = activeSnapshot.Versi

	pkIndividus, err := s.Repo.FindPkIndividus(ctx, snapshot)
	if err != nil {
		return web.RekinPenetapanIndividuResponse{}, err
	}
	pkIndividuIds := make([]int64, 0, len(pkIndividus))
	for _, pk := range pkIndividus {
		pkIndividuIds = append(pkIndividuIds, pk.Id)
	}
	indikators, err := s.Repo.FindIndikatorPkByPkIds(ctx, pkIndividuIds)
	if err != nil {
		return web.RekinPenetapanIndividuResponse{}, err
	}
	indikatorIds := make([]int64, 0, len(indikators))
	for _, ind := range indikators {
		indikatorIds = append(indikatorIds, ind.Id)
	}
	targets, err := s.Repo.FindTargetPkByIndikatorIds(ctx, indikatorIds)
	if err != nil {
		return web.RekinPenetapanIndividuResponse{}, err
	}
	renaksis, err := s.Repo.FindRenaksiIndividuByPkIds(ctx, pkIndividuIds)
	if err != nil {
		return web.RekinPenetapanIndividuResponse{}, err
	}
	renaksiIds := make([]int64, 0, len(renaksis))
	for _, ren := range renaksis {
		renaksiIds = append(renaksiIds, ren.Id)
	}
	pelaksanaans, err := s.Repo.FindPelaksanaanByRenaksiIndividuIds(ctx, renaksiIds)
	if err != nil {
		return web.RekinPenetapanIndividuResponse{}, err
	}

	// To Response DTO
	// pelaksanaan -> renaksi
	pelaksanaanMap := make(map[int64][]web.PelaksanaanRenaksiPkResponse)
	for _, pel := range pelaksanaans {
		pelaksanaanMap[pel.IdRenaksiIndividu] = append(
			pelaksanaanMap[pel.IdRenaksiIndividu],
			web.PelaksanaanRenaksiPkResponse{
				Id:               pel.Id,
				KodePelaksanaan:  pel.KodePelaksanaan,
				BulanPelaksanaan: pel.Bulan,
				BobotPelaksanaan: pel.Bobot,
			})
	}
	renaksiMap := make(map[int64][]web.RenaksiPkResponse)
	for _, ren := range renaksis {
		renaksiMap[ren.IdPk] = append(
			renaksiMap[ren.IdPk],
			web.RenaksiPkResponse{
				Id:              ren.Id,
				KodeRenaksi:     ren.KodeRenaksi,
				NamaRenaksi:     ren.NamaRencanaAksi,
				UrutanRenaksi:   ren.Urutan,
				AnggaranRenaksi: ren.Anggaran,
				Pelaksanaans:    pelaksanaanMap[ren.Id],
			})
	}
	// target -> indikator
	targetMap := make(map[int64][]web.TargetPkResponse)

	for _, target := range targets {
		targetMap[target.IdIndikatorPk] = append(
			targetMap[target.IdIndikatorPk],
			web.TargetPkResponse{
				Id:                   target.Id,
				KodeTargetPk:         target.KodeTargetPk,
				KodeTargetSasaranOpd: target.KodeTargetSasaranOpd,
				Tahun:                target.Tahun,
				Target:               target.Target,
				Satuan:               target.Satuan,
			})
	}

	// indikator -> pk
	indikatorMap := make(map[int64][]web.IndikatorPkResponse)

	for _, indikator := range indikators {
		indikatorMap[indikator.IdPk] = append(
			indikatorMap[indikator.IdPk],
			web.IndikatorPkResponse{
				Id:                      indikator.Id,
				KodeIndikatorPk:         indikator.KodeIndikatorPk,
				NamaIndikatorPk:         indikator.NamaIndikatorPk,
				KodeIndikatorSasaranOpd: indikator.KodeIndikatorSasaranOpd,
				RumusPerhitungan:        indikator.RumusPerhitungan,
				SumberData:              indikator.SumberData,
				DefinisiOperasional:     indikator.DefinisiOperasional,
				TargetPkList:            targetMap[indikator.Id],
			},
		)
	}

	// pk -> response
	rekins := make([]web.RekinIndividuResponse, 0, len(pkIndividus))

	for _, pk := range pkIndividus {
		rekins = append(rekins, web.RekinIndividuResponse{
			Id:              pk.Id,
			LevelPk:         pk.LevelPk,
			KodeSasaranOpd:  pk.KodeSasaranOpd,
			KodePk:          pk.KodePk,
			Rekin:           pk.NamaPk,
			KeteranganPk:    pk.KeteranganPk,
			NamaPemilikPk:   pk.NamaPemilikPk,
			Versi:           pk.Versi,
			AnggaranPk:      pk.AnggaranPk,
			IndikatorPkList: indikatorMap[pk.Id],
			Renaksis:        renaksiMap[pk.Id],
		})
	}

	resp := web.RekinPenetapanIndividuResponse{
		IdPegawai:  snapshot.PegawaiId,
		KodeOpd:    snapshot.KodeOpd,
		TahunAktif: snapshot.Tahun,
		Rekins:     rekins,
	}

	// untuk cek apkaah request pegawaiId sesuai dengan nama pemilik
	if len(pkIndividus) > 0 {
		resp.Nama = pkIndividus[0].NamaPemilikPk
	}

	return resp, nil
}

func (s *PenetapanIndividuService) SyncPenetapanPkIndividu(
	ctx context.Context,
	req *web.SyncPenetapanRequest,
) (web.SyncPenetapanResponse, error) {
	now := time.Now()
	currentUser := "super_admin" // ambil dari ctx nanti
	jenisPenetapan := domain.JenisPenetapanPk

	executor, err := s.SyncExecutor.Get(jenisPenetapan)
	if err != nil {
		s.Logger.Warn("ERROR-SYNC", "error", err, "service", "perencanaan", "jenisPenetapan", jenisPenetapan)
		return web.SyncPenetapanResponse{}, err
	}

	// insert metadata pertama
	metadata := domain.SyncPenetapanMetadata{
		PegawaiId:      req.PegawaiId,
		KodeOpd:        req.KodeOpd,
		Tahun:          req.Tahun,
		Status:         domain.SyncStatusPending,
		StartedAt:      now,
		SyncBy:         &currentUser,
		JenisPenetapan: jenisPenetapan,
	}

	// log start
	s.Logger.InfoContext(
		ctx,
		"sync penetapan pk individu started",
		"pegawai_id", metadata.PegawaiId,
		"kode_opd", metadata.KodeOpd,
		"tahun", metadata.Tahun,
		"jenis_sync", metadata.JenisPenetapan,
		"sync_by", metadata.SyncBy,
		"started_at", metadata.StartedAt,
	)

	syncId, err := s.Repo.InsertMetadata(ctx, metadata)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return web.SyncPenetapanResponse{}, err
	}

	// updateMetadata jadi InProgress
	err = s.markSyncAsInProgress(ctx, syncId)
	if err != nil {
		return web.SyncPenetapanResponse{}, err
	}

	// SYNC AND SAVE
	summary, err := executor.Sync(ctx, syncId, req, currentUser)
	if err != nil {
		return web.SyncPenetapanResponse{}, s.failSync(ctx, syncId, err)
	}

	// jika berhasil updateMetadata jadi Success
	err = s.markSyncAsSuccess(ctx, syncId)
	if err != nil {
		return web.SyncPenetapanResponse{}, err
	}

	return web.SyncPenetapanResponse{
		SyncId:           syncId,
		Status:           domain.SyncStatusSuccess,
		PegawaiId:        req.PegawaiId,
		KodeOpd:          req.KodeOpd,
		Tahun:            req.Tahun,
		JenisPenetapan:   jenisPenetapan,
		ProcessedAt:      time.Now(),
		ProcessedSummary: summary,
	}, nil
}

func (s *PenetapanIndividuService) markSyncAsFailed(ctx context.Context, syncId int64, msg string) error {
	statusSync := domain.SyncStatusFailed
	errorMessage := &msg
	return s.Repo.UpdateMetadataStatus(
		ctx,
		syncId,
		statusSync,
		errorMessage,
	)
}

func (s *PenetapanIndividuService) markSyncAsInProgress(ctx context.Context, syncId int64) error {
	statusSync := domain.SyncStatusInProgress
	return s.Repo.UpdateMetadataStatus(
		ctx,
		syncId,
		statusSync,
		nil,
	)
}

func (s *PenetapanIndividuService) markSyncAsSuccess(ctx context.Context, syncId int64) error {
	statusSync := domain.SyncStatusSuccess
	return s.Repo.UpdateMetadataStatus(
		ctx,
		syncId,
		statusSync,
		nil,
	)
}

func (s *PenetapanIndividuService) failSync(ctx context.Context, syncId int64, cause error) error {
	updateErr := s.markSyncAsFailed(
		ctx,
		syncId,
		cause.Error(),
	)
	if updateErr != nil {
		return updateErr
	}
	return cause
}

func (s *PenetapanIndividuService) getActiveSnapshot(
	ctx context.Context,
	snap domain.SnapshotPenetapan,
) (*domain.ActiveSnapshot, error) {

	snapshot, err := s.Repo.GetActiveSnapshot(ctx,
		snap.PegawaiId,
		snap.KodeOpd,
		snap.Tahun,
		snap.JenisSnapshot,
	)
	if err != nil {
		s.Logger.Error("GetActiveSnapshot")
		return nil, err
	}
	if snapshot == nil {
		s.Logger.WarnContext(
			ctx,
			"get active snapshot individu failed",
			"pegawaiId", snap.PegawaiId,
			"kode_opd", snap.KodeOpd,
			"jenis", snap.JenisSnapshot,
			"tahun", snap.Tahun,
			"err", err,
		)
		return &domain.ActiveSnapshot{}, nil
	}

	return snapshot, nil
}

func (s *PenetapanIndividuService) SyncRenjaIndividu(
	ctx context.Context,
	req *web.SyncPenetapanRequest,
) (web.SyncPenetapanResponse, error) {
	now := time.Now()
	currentUser := "super_admin" // ambil dari ctx nanti
	jenisPenetapan := domain.JenisPenetapanRenjaIndividu

	executor, err := s.SyncExecutor.Get(jenisPenetapan)
	if err != nil {
		return web.SyncPenetapanResponse{}, err
	}

	// insert metadata pertama
	metadata := domain.SyncPenetapanMetadata{
		PegawaiId:      req.PegawaiId,
		KodeOpd:        req.KodeOpd,
		Tahun:          req.Tahun,
		Status:         domain.SyncStatusPending,
		StartedAt:      now,
		SyncBy:         &currentUser,
		JenisPenetapan: jenisPenetapan,
	}

	// log start
	s.Logger.InfoContext(
		ctx,
		"sync penetapan renja individu started",
		"pegawai_id", metadata.PegawaiId,
		"kode_opd", metadata.KodeOpd,
		"tahun", metadata.Tahun,
		"jenis_sync", metadata.JenisPenetapan,
		"sync_by", metadata.SyncBy,
		"started_at", metadata.StartedAt,
	)

	syncId, err := s.Repo.InsertMetadata(ctx, metadata)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return web.SyncPenetapanResponse{}, err
	}

	// updateMetadata jadi InProgress
	err = s.markSyncAsInProgress(ctx, syncId)
	if err != nil {
		return web.SyncPenetapanResponse{}, err
	}

	// SYNC AND SAVE
	summary, err := executor.Sync(ctx, syncId, req, currentUser)
	if err != nil {
		return web.SyncPenetapanResponse{}, s.failSync(ctx, syncId, err)
	}

	// jika berhasil updateMetadata jadi Success
	err = s.markSyncAsSuccess(ctx, syncId)
	if err != nil {
		return web.SyncPenetapanResponse{}, err
	}

	return web.SyncPenetapanResponse{
		SyncId:           syncId,
		Status:           domain.SyncStatusSuccess,
		PegawaiId:        req.PegawaiId,
		KodeOpd:          req.KodeOpd,
		Tahun:            req.Tahun,
		JenisPenetapan:   jenisPenetapan,
		ProcessedAt:      time.Now(),
		ProcessedSummary: summary,
	}, nil
}

func (s *PenetapanIndividuService) FindRenjaIndividu(
	ctx context.Context,
	req web.SyncPenetapanRequest,
) (web.RenjaIndividuResponse, error) {
	s.Logger.Info("FindRenjaIndividu")

	snapshot := domain.SnapshotPenetapan{
		JenisSnapshot: domain.JenisPenetapanRenjaIndividu,
		PegawaiId:     req.PegawaiId,
		KodeOpd:       req.KodeOpd,
		Tahun:         req.Tahun,
	}
	activeSnapshot, err := s.getActiveSnapshot(ctx, snapshot)
	if err != nil {
		return web.RenjaIndividuResponse{}, err
	}
	snapshot.SnapshotId = &activeSnapshot.Id
	snapshot.Versi = activeSnapshot.Versi

	renjas, err := s.Repo.FindRenjaIndividu(ctx, snapshot)
	if err != nil {
		return web.RenjaIndividuResponse{}, err
	}
	renjaIds := make([]int64, 0, len(renjas))
	for _, r := range renjas {
		renjaIds = append(renjaIds, r.Id)
	}

	indikators, err := s.Repo.FindIndikatorRenjaByRenjaIds(ctx, renjaIds)
	if err != nil {
		return web.RenjaIndividuResponse{}, err
	}

	indikatorIds := make([]int64, 0, len(indikators))
	for _, i := range indikators {
		indikatorIds = append(indikatorIds, i.ID)
	}

	targets, err := s.Repo.FindTargetRenjaByIndikatorIds(ctx, indikatorIds)
	if err != nil {
		return web.RenjaIndividuResponse{}, err
	}
	targetMap := make(map[int64][]domain.TargetRenjaIndividu)

	for _, t := range targets {
		targetMap[t.IndikatorRenjaIndividuID] =
			append(targetMap[t.IndikatorRenjaIndividuID], t)
	}

	type indikatorGroup struct {
		Programs     []web.IndikatorRenja
		Kegiatans    []web.IndikatorRenja
		Subkegiatans []web.IndikatorRenja
	}

	indikatorMap := map[int64]indikatorGroup{}

	for _, ind := range indikators {
		var targets []web.TargetRenja

		for _, t := range targetMap[ind.ID] {
			targets = append(targets, web.TargetRenja{
				Id:                 t.ID,
				KodeTarget:         t.KodeTargetRenja,
				KodeTargetRenjaOpd: t.KodeTargetRenjaOpd,
				Target:             t.Target,
				Satuan:             t.Satuan,
				Tahun:              t.Tahun,
			})
		}

		dto := web.IndikatorRenja{
			Id:                    ind.ID,
			KodeIndikator:         ind.KodeIndikatorRenja,
			KodeIndikatorRenjaOpd: ind.KodeIndikatorRenjaOpd,
			Indikator:             ind.Indikator,
			Targets:               targets,
		}

		group := indikatorMap[ind.RenjaIndividuID]

		switch ind.JenisIndikator {
		case "PROGRAM":
			group.Programs = append(group.Programs, dto)

		case "KEGIATAN":
			group.Kegiatans = append(group.Kegiatans, dto)

		case "SUB-KEGIATAN":
			group.Subkegiatans = append(group.Subkegiatans, dto)
		}

		indikatorMap[ind.RenjaIndividuID] = group
	}

	// To Response DTO
	// pk -> response
	renjaIndividus := make([]web.RenjaIndividu, 0, len(renjas))
	for _, renja := range renjas {
		group := indikatorMap[renja.Id]

		renjaIndividus = append(renjaIndividus, web.RenjaIndividu{
			Id:          renja.Id,
			LevelPk:     renja.LevelPk,
			KodePk:      renja.KodePk,
			IdPegawai:   renja.PegawaiId,
			NamaPegawai: renja.NamaPemilikPk,

			KodeProgram:         renja.KodeProgram,
			NamaProgram:         renja.NamaProgram,
			KodePaguProgram:     renja.KodePaguProgram,
			PaguProgram:         renja.PaguProgram,
			IndikatorPrograms:   group.Programs,
			KodeRenjaOpdProgram: renja.KodeRenjaOpdProgram,

			KodeKegiatan:         renja.KodeKegiatan,
			NamaKegiatan:         renja.NamaKegiatan,
			KodePaguKegiatan:     renja.KodePaguKegiatan,
			PaguKegiatan:         renja.PaguKegiatan,
			IndikatorKegiatans:   group.Kegiatans,
			KodeRenjaOpdKegiatan: renja.KodeRenjaOpdKegiatan,

			KodeSubkegiatan:         renja.KodeSubkegiatan,
			NamaSubkegiatan:         renja.NamaSubkegiatan,
			KodePaguSubkegiatan:     renja.KodePaguSubkegiatan,
			PaguSubkegiatan:         renja.PaguSubkegiatan,
			IndikatorSubkegiatans:   group.Subkegiatans,
			KodeRenjaOpdSubkegiatan: renja.KodeRenjaOpdSubkegiatan,
		})
	}

	resp := web.RenjaIndividuResponse{
		IdPegawai:  snapshot.PegawaiId,
		KodeOpd:    snapshot.KodeOpd,
		TahunAktif: snapshot.Tahun,
		Renjas:     renjaIndividus,
	}

	if len(renjas) > 0 {
		resp.Nama = renjas[0].NamaPemilikPk
	}

	return resp, nil
}

func (s *PenetapanIndividuService) FindRenaksiIndividu(
	ctx context.Context,
	req web.SyncPenetapanRequest,
) (web.RenaksiPenetapanIndividuResponse, error) {
	s.Logger.Info("FindRenaksiIndividu")

	snapshot := domain.SnapshotPenetapan{
		JenisSnapshot: domain.JenisPenetapanPk,
		PegawaiId:     req.PegawaiId,
		KodeOpd:       req.KodeOpd,
		Tahun:         req.Tahun,
		Bulan:         req.Bulan,
	}
	activeSnapshot, err := s.getActiveSnapshot(ctx, snapshot)
	if err != nil {
		return web.RenaksiPenetapanIndividuResponse{}, err
	}
	snapshot.SnapshotId = &activeSnapshot.Id
	snapshot.Versi = activeSnapshot.Versi

	pkIndividus, err := s.Repo.FindPkIndividus(ctx, snapshot)
	if err != nil {
		return web.RenaksiPenetapanIndividuResponse{}, err
	}
	pkIndividuIds := make([]int64, 0, len(pkIndividus))
	for _, pk := range pkIndividus {
		pkIndividuIds = append(pkIndividuIds, pk.Id)
	}
	renaksis, err := s.Repo.FindRenaksiIndividuByPkIds(ctx, pkIndividuIds)
	if err != nil {
		return web.RenaksiPenetapanIndividuResponse{}, err
	}
	renaksiIds := make([]int64, 0, len(renaksis))
	for _, ren := range renaksis {
		renaksiIds = append(renaksiIds, ren.Id)
	}
	pelaksanaans, err := s.Repo.FindPelaksanaanByRenaksiIndividuIdsAndBulan(ctx, renaksiIds, snapshot.Bulan)
	if err != nil {
		return web.RenaksiPenetapanIndividuResponse{}, err
	}

	// To Response DTO
	// pelaksanaan -> renaksi
	pelaksanaanMap := make(map[int64][]web.PelaksanaanRenaksiPkResponse)
	for _, pel := range pelaksanaans {
		pelaksanaanMap[pel.IdRenaksiIndividu] = append(
			pelaksanaanMap[pel.IdRenaksiIndividu],
			web.PelaksanaanRenaksiPkResponse{
				Id:               pel.Id,
				KodePelaksanaan:  pel.KodePelaksanaan,
				BulanPelaksanaan: pel.Bulan,
				BobotPelaksanaan: pel.Bobot,
			})
	}
	renaksiMap := make(map[int64][]web.RenaksiPkResponse)
	for _, ren := range renaksis {
		renaksiMap[ren.IdPk] = append(
			renaksiMap[ren.IdPk],
			web.RenaksiPkResponse{
				Id:              ren.Id,
				KodeRenaksi:     ren.KodeRenaksi,
				NamaRenaksi:     ren.NamaRencanaAksi,
				UrutanRenaksi:   ren.Urutan,
				AnggaranRenaksi: ren.Anggaran,
				Pelaksanaans:    pelaksanaanMap[ren.Id],
			})
	}

	// pk -> response
	rekins := make([]web.RekinIndividuResponse, 0, len(pkIndividus))

	for _, pk := range pkIndividus {
		rekins = append(rekins, web.RekinIndividuResponse{
			Id:             pk.Id,
			LevelPk:        pk.LevelPk,
			KodeSasaranOpd: pk.KodeSasaranOpd,
			KodePk:         pk.KodePk,
			Rekin:          pk.NamaPk,
			KeteranganPk:   pk.KeteranganPk,
			NamaPemilikPk:  pk.NamaPemilikPk,
			Versi:          pk.Versi,
			AnggaranPk:     pk.AnggaranPk,
			Renaksis:       renaksiMap[pk.Id],
		})
	}

	resp := web.RenaksiPenetapanIndividuResponse{
		IdPegawai:  snapshot.PegawaiId,
		KodeOpd:    snapshot.KodeOpd,
		TahunAktif: snapshot.Tahun,
		Bulan:      snapshot.Bulan,
		Rekins:     rekins,
	}

	// untuk cek apkaah request pegawaiId sesuai dengan nama pemilik
	if len(pkIndividus) > 0 {
		resp.Nama = pkIndividus[0].NamaPemilikPk
	}

	return resp, nil
}

func (s *PenetapanIndividuService) UpdatePenetapanPkIndividu(
	ctx context.Context,
	req *web.SyncPenetapanRequest,
) error {
	// db transaction
	tx, err := s.Repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// find snapshot
	snapshot := domain.SnapshotPenetapan{
		JenisSnapshot: domain.JenisPenetapanPk,
		PegawaiId:     req.PegawaiId,
		KodeOpd:       req.KodeOpd,
		Tahun:         req.Tahun,
	}
	activeSnapshot, err := s.getActiveSnapshot(ctx, snapshot)
	if err != nil {
		return err
	}
	snapshot.Id = activeSnapshot.Id
	log.Printf("SNAPSHOT: %v", snapshot)

	err = s.Repo.DeactivateOldSnapshot(ctx, tx, snapshot)
	if err != nil {
		return err
	}
	// commit transaction penetapan tujuan
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
