-- +goose Up
-- +goose StatementBegin
CREATE TABLE aircraft_data (
    id SERIAL PRIMARY KEY,
    icao_code VARCHAR(20),
    faa_designator VARCHAR(30),
    manufacturer VARCHAR(100),
    model_faa VARCHAR(100),
    model_bada VARCHAR(100),
    physical_class_engine VARCHAR(30),
    num_engines INTEGER,
    aac VARCHAR(20),
    aac_minimum VARCHAR(20),
    aac_maximum VARCHAR(20),
    adg VARCHAR(20),
    tdg VARCHAR(20),
    approach_speed_knot INTEGER,
    approach_speed_minimum_knot INTEGER,
    approach_speed_maximum_knot INTEGER,
    wingspan_ft_without_winglets_sharklets DECIMAL(8,2),
    wingspan_ft_with_winglets_sharklets DECIMAL(8,2),
    length_ft DECIMAL(8,2),
    tail_height_at_oew_ft DECIMAL(8,2),
    wheelbase_ft DECIMAL(8,2),
    cockpit_to_main_gear_ft DECIMAL(8,2),
    main_gear_width_ft DECIMAL(8,2),
    mtow_lb INTEGER,
    malw_lb INTEGER,
    main_gear_config VARCHAR(20),
    icao_wtc VARCHAR(30),
    parking_area_ft2 DECIMAL(10,2),
    class VARCHAR(30),
    faa_weight VARCHAR(20),
    cwt VARCHAR(20),
    one_half_wake_category VARCHAR(20),
    two_wake_category_appx_a VARCHAR(20),
    two_wake_category_appx_b VARCHAR(20),
    rotor_diameter_ft DECIMAL(8,2),
    srs VARCHAR(20),
    lahso VARCHAR(20),
    faa_registry VARCHAR(20),
    registration_count INTEGER,
    tmfs_operations_fy24 INTEGER,
    remarks TEXT,
    last_update VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create unique constraint for upsert functionality
ALTER TABLE aircraft_data ADD CONSTRAINT uk_aircraft_codes UNIQUE (icao_code, faa_designator);

-- Create indexes for common search fields
CREATE INDEX idx_aircraft_icao_code ON aircraft_data(icao_code);
CREATE INDEX idx_aircraft_faa_designator ON aircraft_data(faa_designator);
CREATE INDEX idx_aircraft_manufacturer ON aircraft_data(manufacturer);
CREATE INDEX idx_aircraft_model_faa ON aircraft_data(model_faa);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS aircraft_data;
-- +goose StatementEnd
