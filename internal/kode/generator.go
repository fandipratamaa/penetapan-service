package kode

import (
	"fmt"
	"strings"
)

func KodeSasaranOpd(sasaranId string) string {
	kodeSasaran := fmt.Sprintf("SAS-OPD-%s", sasaranId)

	return kodeSasaran
}

func KodeSasaranOpdInt(sasaranId int) string {
	kodeSasaran := fmt.Sprintf("SAS-OPD-%d", sasaranId)

	return kodeSasaran
}

func KodeIndikatorSasaranOpd(indikatorSasaranId string) string {
	if strings.TrimSpace(indikatorSasaranId) == "" {
		return ""
	}
	return fmt.Sprintf("IND-%s", indikatorSasaranId)
}

func KodeTargetSasaranOpd(targetSasaranId string) string {
	if strings.TrimSpace(targetSasaranId) == "" {
		return ""
	}
	return fmt.Sprintf("TGT-%s", targetSasaranId)
}

func KodeRenaksiOpd(renaksiId int) string {
	return fmt.Sprintf("REN-OPD-%d", renaksiId)
}

// KodeRenjaOpd membuat kode induk renja untuk satu tier cascade
// (program / kegiatan / subkegiatan).
//
// Parameter sengaja int64, bukan string, supaya "tidak ditemukan" (0) tidak
// bisa berubah jadi kode "RNJA-OPD-0" di pemanggil. Banyak renja OPD belum
// punya baris di tabel cascade-nya, dan kode palsu lebih berbahaya dari kode
// kosong karena terlihat sahih.
func KodeRenjaOpd(renjaOpdId int64) string {
	if renjaOpdId <= 0 {
		return ""
	}
	return fmt.Sprintf("RNJA-OPD-%d", renjaOpdId)
}

func KodeIndikatorRenjaOpd(indikatorRenjaOpdId int64) string {
	if indikatorRenjaOpdId <= 0 {
		return ""
	}
	return fmt.Sprintf("IND-%d", indikatorRenjaOpdId)
}

func KodeTargetRenjaOpd(targetRenjaOpdId int64) string {
	if targetRenjaOpdId <= 0 {
		return ""
	}
	return fmt.Sprintf("TGT-%d", targetRenjaOpdId)
}
