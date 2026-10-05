package domain

import (
	"time"
)

type RenjaIndividu struct {
	Id int64

	PenetapanIndividuId int64
	KodePk              string
	LevelPk             int
	NamaPemilikPk       string

	PegawaiId  string
	KodeOpd    string
	TahunAktif int

	KodeProgram       string
	NamaProgram       string
	KodePaguProgram   string
	PaguProgram       int64
	IndikatorPrograms []IndikatorRenjaIndividu

	KodeKegiatan       string
	NamaKegiatan       string
	KodePaguKegiatan   string
	PaguKegiatan       int64
	IndikatorKegiatans []IndikatorRenjaIndividu

	KodeSubkegiatan       string
	NamaSubkegiatan       string
	KodePaguSubkegiatan   string
	PaguSubkegiatan       int64
	IndikatorSubkegiatans []IndikatorRenjaIndividu

	KodeRenjaOpdProgram     string
	KodeRenjaOpdKegiatan    string
	KodeRenjaOpdSubkegiatan string

	CreatedDate      time.Time
	LastModifiedDate time.Time
	CreatedBy        *string
}

type IndikatorRenjaIndividu struct {
	ID                    int64
	RenjaIndividuID       int64
	JenisIndikator        string
	KodeIndikatorRenja    string
	Indikator             string
	KodeIndikatorRenjaOpd string
	Targets               []TargetRenjaIndividu
}

type TargetRenjaIndividu struct {
	ID                       int64
	IndikatorRenjaIndividuID int64
	JenisTarget              string
	KodeTargetRenja          string
	KodeTargetRenjaOpd       string
	Target                   float64
	Satuan                   string
	Tahun                    int
}
