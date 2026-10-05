package web

type RenjaIndividuResponse struct {
	IdPegawai  string          `json:"pegawai_id" example:"12345777"`
	Nama       string          `json:"nama" example:"Pegawai X"`
	KodeOpd    string          `json:"kode_opd" example:"1.23.456"`
	TahunAktif int             `json:"tahun_aktif" example:"2025"`
	Renjas     []RenjaIndividu `json:"renjas"`
}

type RenjaIndividu struct {
	Id          int64  `json:"id" example:"1"`
	KodePk      string `json:"kode_pk" example:"RK-001"`
	LevelPk     int    `json:"level_pk" example:"5"`
	IdPegawai   string `json:"pegawai_id" example:"12345"`
	NamaPegawai string `json:"nama_pegawai" example:"Pegawai X"`

	KodeProgram         string           `json:"kode_program" example:"8.01.05"`
	NamaProgram         string           `json:"nama_program" example:"PROGRAM ABC"`
	KodePaguProgram     string           `json:"kode_pagu_program" example:"PAGU-PRG-xxxx"`
	PaguProgram         int64            `json:"pagu_program" example:"2500000000"`
	IndikatorPrograms   []IndikatorRenja `json:"indikator_programs"`
	KodeRenjaOpdProgram string           `json:"kode_renja_opd_program" example:"RNJA-OPD-1"`

	KodeKegiatan         string           `json:"kode_kegiatan" example:"8.01.05.2.01"`
	NamaKegiatan         string           `json:"nama_kegiatan" example:"KEGIATAN XX"`
	KodePaguKegiatan     string           `json:"kode_pagu_kegiatan" example:"PAGU-KEG-xxxx"`
	PaguKegiatan         int64            `json:"pagu_kegiatan" example:"1250000000"`
	IndikatorKegiatans   []IndikatorRenja `json:"indikator_kegiatans"`
	KodeRenjaOpdKegiatan string           `json:"kode_renja_opd_kegiatan" example:"RNJA-OPD-2"`

	KodeSubkegiatan         string           `json:"kode_subkegiatan" example:"8.01.05.2.01.0003"`
	NamaSubkegiatan         string           `json:"nama_subkegiatan" example:"SUBKEGIATAN X"`
	KodePaguSubkegiatan     string           `json:"kode_pagu_subkegiatan" example:"PAGU-SUBKEG-xxxx"`
	PaguSubkegiatan         int64            `json:"pagu_subkegiatan" example:"500000000"`
	IndikatorSubkegiatans   []IndikatorRenja `json:"indikator_subkegiatans"`
	KodeRenjaOpdSubkegiatan string           `json:"kode_renja_opd_subkegiatan" example:"RNJA-OPD-3"`
}

type IndikatorRenja struct {
	Id                    int64         `json:"id" example:"1"`
	KodeIndikatorRenjaOpd string        `json:"kode_indikator_renja_opd" example:"IND-5"`
	KodeIndikator         string        `json:"kode_indikator" example:"IND-RENJA-123"`
	Indikator             string        `json:"indikator" example:"INDIKATOR-PROGRAM-1"`
	Targets               []TargetRenja `json:"targets"`
}

type TargetRenja struct {
	Id                 int64   `json:"id" example:"1"`
	KodeTargetRenjaOpd string  `json:"kode_target_renja_opd" example:"TGT-9"`
	KodeTarget         string  `json:"kode_target" example:"TRG-RENJA-123"`
	Target             float64 `json:"target" example:"12"`
	Satuan             string  `json:"satuan" example:"%"`
	Tahun              int     `json:"tahun" example:"2026"`
}
