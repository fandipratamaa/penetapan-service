package web

type RekinPenetapanIndividuResponse struct {
	IdPegawai  string                  `json:"pegawai_id" example:"12345"`
	Nama       string                  `json:"nama" example:"Pegawai X"`
	KodeOpd    string                  `json:"kode_opd" example:"1.23.456"`
	TahunAktif int                     `json:"tahun_aktif" example:"2025"`
	Rekins     []RekinIndividuResponse `json:"rekins"`
}

type RenaksiPenetapanIndividuResponse struct {
	IdPegawai  string                  `json:"pegawai_id" example:"12345"`
	Nama       string                  `json:"nama" example:"Pegawai X"`
	KodeOpd    string                  `json:"kode_opd" example:"1.23.456"`
	TahunAktif int                     `json:"tahun_aktif" example:"2025"`
	Bulan      int                     `json:"bulan" example:"10"`
	Rekins     []RekinIndividuResponse `json:"rekins"`
}

type RekinIndividuResponse struct {
	Id              int64                 `json:"id" example:"1"`
	LevelPk         int                   `json:"level_pk" example:"2"`
	KodeSasaranOpd  string                `json:"kode_sasaran_opd" example:"SAS-OPD-1"`
	KodePk          string                `json:"kode_pk" example:"RK-001"`
	Rekin           string                `json:"rekin" example:"Terwujudnya pelayanan publik yang berkualitas"`
	KeteranganPk    string                `json:"keterangan_pk,omitempty" example:"Kinerja utama individu"`
	NamaPemilikPk   string                `json:"nama_pemilik_pk" example:"Budi Santoso"`
	AnggaranPk      int                   `json:"anggaran_pk" example:"500000"`
	Versi           int                   `json:"versi" example:"1"`
	IndikatorPkList []IndikatorPkResponse `json:"indikator_pk"`
	Renaksis        []RenaksiPkResponse   `json:"renaksis"`
}

type IndikatorPkResponse struct {
	Id                      int64              `json:"id" example:"1"`
	KodeIndikatorSasaranOpd string             `json:"kode_indikator_sasaran_opd" example:"IND-1"`
	KodeIndikatorPk         string             `json:"kode_indikator_pk" example:"IKU-001"`
	NamaIndikatorPk         string             `json:"nama_indikator_pk" example:"Persentase kepuasan masyarakat"`
	RumusPerhitungan        *string            `json:"rumus_perhitungan,omitempty" example:"Jumlah puas / total responden x 100%"`
	SumberData              *string            `json:"sumber_data,omitempty" example:"Survei Kepuasan Masyarakat"`
	DefinisiOperasional     *string            `json:"definisi_operasional,omitempty" example:"Persentase responden yang menyatakan puas"`
	TargetPkList            []TargetPkResponse `json:"target_pk"`
}

type TargetPkResponse struct {
	Id                   int64   `json:"id" example:"1"`
	KodeTargetSasaranOpd string  `json:"kode_target_sasaran_opd" example:"TGT-1"`
	KodeTargetPk         string  `json:"kode_target_pk" example:"TGT-001"`
	Tahun                int     `json:"tahun" example:"2025"`
	Target               float64 `json:"target" example:"95"`
	Satuan               string  `json:"satuan" example:"persen"`
}

type RenaksiPkResponse struct {
	Id              int64                          `json:"id" example:"1"`
	UrutanRenaksi   int                            `json:"urutan_renaksi" example:"1"`
	KodeRenaksi     string                         `json:"kode_renaksi" example:"REN-001"`
	NamaRenaksi     string                         `json:"nama_renaksi" example:"Rapat Koordinasi"`
	AnggaranRenaksi int                            `json:"anggaran_renaksi" example:"100000"`
	Pelaksanaans    []PelaksanaanRenaksiPkResponse `json:"pelaksanaans"`
}

type PelaksanaanRenaksiPkResponse struct {
	Id               int64  `json:"id" example:"1"`
	KodePelaksanaan  string `json:"kode_pelaksanaan" example:"PEL-001"`
	BulanPelaksanaan int    `json:"bulan_pelaksanaan" example:"12"`
	BobotPelaksanaan int    `json:"bobot_pelaksanaan" example:"10"`
}
