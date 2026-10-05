package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/bappeda-dev-team/penetapan-service/internal/individu/domain"
)

func (repo *PenetapanIndividuRepository) SaveRenjaIndividu(
	ctx context.Context,
	tx *sql.Tx,
	req domain.RenjaIndividu,
) (int64, error) {
	const query = `
		INSERT INTO renja_individu (
			penetapan_individu_id,
			kode_pk,
			level_pk,
			pegawai_id,
			nama_pemilik_pk,
			kode_opd,
			tahun_aktif,
			kode_program,
			nama_program,
			kode_pagu_program,
			pagu_program,
			kode_kegiatan,
			nama_kegiatan,
			kode_pagu_kegiatan,
			pagu_kegiatan,
			kode_subkegiatan,
			nama_subkegiatan,
			kode_pagu_subkegiatan,
			pagu_subkegiatan,
			kode_renja_opd_program,
			kode_renja_opd_kegiatan,
			kode_renja_opd_subkegiatan,
			created_by
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18,
			$19, $20, $21, $22,
			$23
		)
		RETURNING id
	`

	var id int64

	err := tx.QueryRowContext(
		ctx,
		query,
		req.PenetapanIndividuId,
		req.KodePk,
		req.LevelPk,
		req.PegawaiId,
		req.NamaPemilikPk,
		req.KodeOpd,
		req.TahunAktif,
		req.KodeProgram,
		req.NamaProgram,
		req.KodePaguProgram,
		req.PaguProgram,

		req.KodeKegiatan,
		req.NamaKegiatan,
		req.KodePaguKegiatan,
		req.PaguKegiatan,

		req.KodeSubkegiatan,
		req.NamaSubkegiatan,
		req.KodePaguSubkegiatan,
		req.PaguSubkegiatan,
		req.KodeRenjaOpdProgram,
		req.KodeRenjaOpdKegiatan,
		req.KodeRenjaOpdSubkegiatan,
		req.CreatedBy,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *PenetapanIndividuRepository) SaveIndikatorRenjaIndividu(
	ctx context.Context,
	tx *sql.Tx,
	req domain.IndikatorRenjaIndividu,
) (int64, error) {
	const query = `
		INSERT INTO indikator_renja_individu (
			renja_individu_id,
			jenis_indikator,
			kode_indikator_renja,
			indikator,
			kode_indikator_renja_opd
		) VALUES (
			$1, $2, $3, $4, $5
		)
		RETURNING id
	`

	var id int64

	err := tx.QueryRowContext(
		ctx,
		query,
		req.RenjaIndividuID,
		req.JenisIndikator,
		req.KodeIndikatorRenja,
		req.Indikator,
		req.KodeIndikatorRenjaOpd,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *PenetapanIndividuRepository) SaveTargetRenjaIndividu(
	ctx context.Context,
	tx *sql.Tx,
	req domain.TargetRenjaIndividu,
) (int64, error) {
	const query = `
		INSERT INTO target_renja_individu (
			indikator_renja_individu_id,
			jenis_target,
			kode_target_renja,
			target,
			satuan,
			tahun,
			kode_target_renja_opd
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING id
	`

	var id int64

	err := tx.QueryRowContext(
		ctx,
		query,
		req.IndikatorRenjaIndividuID,
		req.JenisTarget,
		req.KodeTargetRenja,
		req.Target,
		req.Satuan,
		req.Tahun,
		req.KodeTargetRenjaOpd,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *PenetapanIndividuRepository) FindRenjaIndividu(
	ctx context.Context,
	req domain.SnapshotPenetapan,
) ([]domain.RenjaIndividu, error) {

	const query = `
		SELECT
			ri.id,
			ri.penetapan_individu_id,
			ri.level_pk,
			ri.kode_pk,
			ri.nama_pemilik_pk,
			ri.pegawai_id,
			ri.kode_opd,
			ri.tahun_aktif,
			ri.kode_program,
			ri.nama_program,
			ri.kode_pagu_program,
			ri.pagu_program,
			ri.kode_kegiatan,
			ri.nama_kegiatan,
			ri.kode_pagu_kegiatan,
			ri.pagu_kegiatan,
			ri.kode_subkegiatan,
			ri.nama_subkegiatan,
			ri.kode_pagu_subkegiatan,
			ri.pagu_subkegiatan,
			ri.kode_renja_opd_program,
			ri.kode_renja_opd_kegiatan,
			ri.kode_renja_opd_subkegiatan,
			ri.created_date,
			ri.last_modified_date,
			ri.created_by
		FROM renja_individu ri
		WHERE ri.pegawai_id = $1
		AND ri.kode_opd = $2
		AND ri.tahun_aktif = $3
		AND ri.penetapan_individu_id = $4
		ORDER BY
			ri.level_pk ASC,
			ri.kode_program ASC,
			ri.kode_kegiatan ASC,
			ri.kode_subkegiatan ASC
	`

	rows, err := repo.DB.QueryContext(
		ctx,
		query,
		req.PegawaiId,
		req.KodeOpd,
		req.Tahun,
		req.SnapshotId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	renjas := make([]domain.RenjaIndividu, 0)

	for rows.Next() {
		var renja domain.RenjaIndividu

		err := rows.Scan(
			&renja.Id,
			&renja.PenetapanIndividuId,
			&renja.LevelPk,
			&renja.KodePk,
			&renja.NamaPemilikPk,
			&renja.PegawaiId,
			&renja.KodeOpd,
			&renja.TahunAktif,
			&renja.KodeProgram,
			&renja.NamaProgram,
			&renja.KodePaguProgram,
			&renja.PaguProgram,
			&renja.KodeKegiatan,
			&renja.NamaKegiatan,
			&renja.KodePaguKegiatan,
			&renja.PaguKegiatan,
			&renja.KodeSubkegiatan,
			&renja.NamaSubkegiatan,
			&renja.KodePaguSubkegiatan,
			&renja.PaguSubkegiatan,
			&renja.KodeRenjaOpdProgram,
			&renja.KodeRenjaOpdKegiatan,
			&renja.KodeRenjaOpdSubkegiatan,
			&renja.CreatedDate,
			&renja.LastModifiedDate,
			&renja.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		renjas = append(renjas, renja)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return renjas, nil
}

func (repo *PenetapanIndividuRepository) FindIndikatorRenjaByRenjaIds(
	ctx context.Context,
	renjaIds []int64,
) ([]domain.IndikatorRenjaIndividu, error) {
	if len(renjaIds) == 0 {
		return []domain.IndikatorRenjaIndividu{}, nil
	}

	var (
		placeholders []string
		args         []any
	)

	for i, id := range renjaIds {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			renja_individu_id,
			jenis_indikator,
			kode_indikator_renja,
			indikator,
			kode_indikator_renja_opd
		FROM indikator_renja_individu
		WHERE renja_individu_id IN (%s)
		ORDER BY id
	`, strings.Join(placeholders, ","))

	rows, err := repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.IndikatorRenjaIndividu

	for rows.Next() {
		var item domain.IndikatorRenjaIndividu

		err := rows.Scan(
			&item.ID,
			&item.RenjaIndividuID,
			&item.JenisIndikator,
			&item.KodeIndikatorRenja,
			&item.Indikator,
			&item.KodeIndikatorRenjaOpd,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

func (repo *PenetapanIndividuRepository) FindTargetRenjaByIndikatorIds(
	ctx context.Context,
	indikatorIds []int64,
) ([]domain.TargetRenjaIndividu, error) {
	if len(indikatorIds) == 0 {
		return []domain.TargetRenjaIndividu{}, nil
	}

	var (
		placeholders []string
		args         []any
	)

	for i, id := range indikatorIds {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			indikator_renja_individu_id,
			jenis_target,
			kode_target_renja,
			target,
			satuan,
			tahun,
			kode_target_renja_opd
		FROM target_renja_individu
		WHERE indikator_renja_individu_id IN (%s)
		ORDER BY tahun
	`, strings.Join(placeholders, ","))

	rows, err := repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.TargetRenjaIndividu

	for rows.Next() {
		var item domain.TargetRenjaIndividu

		err := rows.Scan(
			&item.ID,
			&item.IndikatorRenjaIndividuID,
			&item.JenisTarget,
			&item.KodeTargetRenja,
			&item.Target,
			&item.Satuan,
			&item.Tahun,
			&item.KodeTargetRenjaOpd,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}
