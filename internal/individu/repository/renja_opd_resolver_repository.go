package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/bappeda-dev-team/penetapan-service/internal/model/domain"
)

const (
	TierRenjaProgram     = "PROGRAM"
	TierRenjaKegiatan    = "KEGIATAN"
	TierRenjaSubkegiatan = "SUB-KEGIATAN"
)

// prefixIndikatorOpd ditambahkan pada kode_indikator di cascade renja OPD
const prefixIndikatorOpd = "IND-"

type RenjaOpdRef struct {
	IdProgram     int64
	IdKegiatan    int64
	IdSubkegiatan int64
}

type IndikatorRenjaOpdRef struct {
	ByKode map[string]int64
	ByNama map[string]int64
}

// tierTable memetakan tier ke nama tabel indikator/target.
func tierTable(tier string) (indikatorTable string, targetTable string, parentFK string, targetFK string, ok bool) {
	switch tier {
	case TierRenjaProgram:
		return "indikator_renja_program", "target_indikator_renja_program",
			"program_id", "indikator_program_id", true
	case TierRenjaKegiatan:
		return "indikator_renja_kegiatan", "target_indikator_renja_kegiatan",
			"kegiatan_id", "indikator_kegiatan_id", true
	case TierRenjaSubkegiatan:
		return "indikator_renja_subkegiatan", "target_indikator_renja_subkegiatan",
			"subkegiatan_id", "indikator_subkegiatan_id", true
	}
	return "", "", "", "", false
}

// NormalizeNama menyamakan normalisasi nama indikator individu dengan normalisasi
// di sisi Postgres (LOWER(TRIM(x)) + collapse whitespace). Dipakai hanya sebagai
// cadangan pencocokan nama.
func NormalizeNama(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

func (repo *PenetapanIndividuRepository) ResolveRenjaOpdIds(
	ctx context.Context,
	tx *sql.Tx,
	kodeOpd string,
	tahun int,
	kodeSubkegiatans []string,
) (map[string]RenjaOpdRef, error) {
	result := make(map[string]RenjaOpdRef)
	if len(kodeSubkegiatans) == 0 {
		return result, nil
	}

	const query = `
		SELECT
			s.kode_subkegiatan,
			s.id,
			k.id,
			p.id
		FROM renja_subkegiatan s
		JOIN penetapan_opd po
			ON po.id = s.penetapan_id
			AND po.jenis_penetapan = $1
			AND po.is_active = true
			AND po.kode_opd = $2
			AND po.tahun = $3
		JOIN renja_kegiatan k
			ON k.kode_kegiatan = s.kode_kegiatan
			AND k.penetapan_id = s.penetapan_id
		JOIN renja_program p
			ON p.kode_program = k.kode_program
			AND p.penetapan_id = s.penetapan_id
		WHERE s.kode_subkegiatan = ANY($4)
	`

	rows, err := tx.QueryContext(ctx, query,
		domain.JenisPenetapanRenja, kodeOpd, tahun, kodeSubkegiatans)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			kode string
			ref  RenjaOpdRef
		)
		if err := rows.Scan(&kode, &ref.IdSubkegiatan, &ref.IdKegiatan, &ref.IdProgram); err != nil {
			return nil, err
		}
		// Satu kode_subkegiatan bisa muncul di beberapa snapshot; ambil yang pertama.
		if _, exists := result[kode]; !exists {
			result[kode] = ref
		}
	}

	return result, rows.Err()
}

func (repo *PenetapanIndividuRepository) ResolveIndikatorRenjaOpd(
	ctx context.Context,
	tx *sql.Tx,
	tier string,
	parentId int64,
	tahun int,
) (IndikatorRenjaOpdRef, error) {
	result := IndikatorRenjaOpdRef{
		ByKode: make(map[string]int64),
		ByNama: make(map[string]int64),
	}

	indikatorTable, _, parentFK, _, ok := tierTable(tier)
	if !ok {
		return result, fmt.Errorf("tier renja tidak dikenal: %q", tier)
	}
	if parentId <= 0 {
		return result, nil
	}

	query := fmt.Sprintf(`
		SELECT id, kode_indikator, indikator
		FROM %s
		WHERE %s = $1 AND tahun = $2
		ORDER BY id
	`, indikatorTable, parentFK)

	rows, err := tx.QueryContext(ctx, query, parentId, tahun)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id            int64
			kodeIndikator string
			nama          string
		)
		if err := rows.Scan(&id, &kodeIndikator, &nama); err != nil {
			return result, err
		}

		if kode := strings.TrimPrefix(kodeIndikator, prefixIndikatorOpd); kode != "" {
			if existing, dup := result.ByKode[kode]; !dup || id < existing {
				result.ByKode[kode] = id
			}
		}

		if key := NormalizeNama(nama); key != "" {
			if existing, dup := result.ByNama[key]; !dup || id < existing {
				result.ByNama[key] = id
			}
		}
	}

	return result, rows.Err()
}

func (r IndikatorRenjaOpdRef) Lookup(kodeIndikator string, nama string) (int64, bool) {
	if id, ok := r.ByKode[strings.TrimSpace(kodeIndikator)]; ok {
		return id, true
	}

	if id, ok := r.ByNama[NormalizeNama(nama)]; ok {
		return id, true
	}

	return 0, false
}

func (repo *PenetapanIndividuRepository) ResolveTargetRenjaOpd(
	ctx context.Context,
	tx *sql.Tx,
	tier string,
	indikatorOpdId int64,
) (map[int]int64, error) {
	result := make(map[int]int64)

	_, targetTable, _, targetFK, ok := tierTable(tier)
	if !ok {
		return result, fmt.Errorf("tier renja tidak dikenal: %q", tier)
	}
	if indikatorOpdId <= 0 {
		return result, nil
	}

	query := fmt.Sprintf(`
		SELECT tahun, id
		FROM %s
		WHERE %s = $1
		ORDER BY id
	`, targetTable, targetFK)

	rows, err := tx.QueryContext(ctx, query, indikatorOpdId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tahun int
			id    int64
		)
		if err := rows.Scan(&tahun, &id); err != nil {
			return nil, err
		}
		// Beberapa target pada tahun yang sama: ambil ID terkecil.
		if existing, dup := result[tahun]; !dup || id < existing {
			result[tahun] = id
		}
	}

	return result, rows.Err()
}
