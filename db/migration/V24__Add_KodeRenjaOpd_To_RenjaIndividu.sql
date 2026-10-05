ALTER TABLE renja_individu
    ADD COLUMN kode_renja_opd_program     VARCHAR(255),
    ADD COLUMN kode_renja_opd_kegiatan    VARCHAR(255),
    ADD COLUMN kode_renja_opd_subkegiatan VARCHAR(255);

ALTER TABLE indikator_renja_individu
    ADD COLUMN kode_indikator_renja_opd VARCHAR(255);

ALTER TABLE target_renja_individu
    ADD COLUMN kode_target_renja_opd VARCHAR(255);
