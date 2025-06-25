-- name: GetAircraft :one
SELECT * FROM aircraft_data
WHERE id = $1 LIMIT 1;

-- name: SearchAircraft :many
SELECT * FROM aircraft_data
WHERE 
    UPPER(icao_code) LIKE UPPER(@search_term::text) OR 
    UPPER(faa_designator) LIKE UPPER(@search_term::text) OR 
    UPPER(manufacturer) LIKE UPPER(@search_term::text) OR 
    UPPER(model_faa) LIKE UPPER(@search_term::text)
ORDER BY manufacturer, model_faa
LIMIT $1 OFFSET $2;

-- name: CountAircraft :one
SELECT COUNT(*) FROM aircraft_data;

-- name: GetAllAircraft :many
SELECT * FROM aircraft_data
ORDER BY manufacturer, model_faa
LIMIT $1 OFFSET $2;

-- name: CountSearchAircraft :one
SELECT COUNT(*) FROM aircraft_data
WHERE 
    UPPER(icao_code) LIKE UPPER(@search_term::text) OR 
    UPPER(faa_designator) LIKE UPPER(@search_term::text) OR 
    UPPER(manufacturer) LIKE UPPER(@search_term::text) OR 
    UPPER(model_faa) LIKE UPPER(@search_term::text);

-- name: CreateAircraftData :one
INSERT INTO aircraft_data (
    icao_code, faa_designator, manufacturer, model_faa, model_bada,
    physical_class_engine, num_engines, aac, aac_minimum, aac_maximum,
    adg, tdg, approach_speed_knot, approach_speed_minimum_knot, approach_speed_maximum_knot,
    wingspan_ft_without_winglets_sharklets, wingspan_ft_with_winglets_sharklets,
    length_ft, tail_height_at_oew_ft, wheelbase_ft, cockpit_to_main_gear_ft,
    main_gear_width_ft, mtow_lb, malw_lb, main_gear_config, icao_wtc,
    parking_area_ft2, class, faa_weight, cwt, one_half_wake_category,
    two_wake_category_appx_a, two_wake_category_appx_b, rotor_diameter_ft,
    srs, lahso, faa_registry, registration_count, tmfs_operations_fy24,
    remarks, last_update
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
    $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41
) RETURNING *;

-- name: UpsertAircraftData :one
INSERT INTO aircraft_data (
    icao_code, faa_designator, manufacturer, model_faa, model_bada,
    physical_class_engine, num_engines, aac, aac_minimum, aac_maximum,
    adg, tdg, approach_speed_knot, approach_speed_minimum_knot, approach_speed_maximum_knot,
    wingspan_ft_without_winglets_sharklets, wingspan_ft_with_winglets_sharklets,
    length_ft, tail_height_at_oew_ft, wheelbase_ft, cockpit_to_main_gear_ft,
    main_gear_width_ft, mtow_lb, malw_lb, main_gear_config, icao_wtc,
    parking_area_ft2, class, faa_weight, cwt, one_half_wake_category,
    two_wake_category_appx_a, two_wake_category_appx_b, rotor_diameter_ft,
    srs, lahso, faa_registry, registration_count, tmfs_operations_fy24,
    remarks, last_update
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
    $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41
)
ON CONFLICT (icao_code, faa_designator) DO UPDATE SET
    manufacturer = EXCLUDED.manufacturer,
    model_faa = EXCLUDED.model_faa,
    model_bada = EXCLUDED.model_bada,
    physical_class_engine = EXCLUDED.physical_class_engine,
    num_engines = EXCLUDED.num_engines,
    aac = EXCLUDED.aac,
    aac_minimum = EXCLUDED.aac_minimum,
    aac_maximum = EXCLUDED.aac_maximum,
    adg = EXCLUDED.adg,
    tdg = EXCLUDED.tdg,
    approach_speed_knot = EXCLUDED.approach_speed_knot,
    approach_speed_minimum_knot = EXCLUDED.approach_speed_minimum_knot,
    approach_speed_maximum_knot = EXCLUDED.approach_speed_maximum_knot,
    wingspan_ft_without_winglets_sharklets = EXCLUDED.wingspan_ft_without_winglets_sharklets,
    wingspan_ft_with_winglets_sharklets = EXCLUDED.wingspan_ft_with_winglets_sharklets,
    length_ft = EXCLUDED.length_ft,
    tail_height_at_oew_ft = EXCLUDED.tail_height_at_oew_ft,
    wheelbase_ft = EXCLUDED.wheelbase_ft,
    cockpit_to_main_gear_ft = EXCLUDED.cockpit_to_main_gear_ft,
    main_gear_width_ft = EXCLUDED.main_gear_width_ft,
    mtow_lb = EXCLUDED.mtow_lb,
    malw_lb = EXCLUDED.malw_lb,
    main_gear_config = EXCLUDED.main_gear_config,
    icao_wtc = EXCLUDED.icao_wtc,
    parking_area_ft2 = EXCLUDED.parking_area_ft2,
    class = EXCLUDED.class,
    faa_weight = EXCLUDED.faa_weight,
    cwt = EXCLUDED.cwt,
    one_half_wake_category = EXCLUDED.one_half_wake_category,
    two_wake_category_appx_a = EXCLUDED.two_wake_category_appx_a,
    two_wake_category_appx_b = EXCLUDED.two_wake_category_appx_b,
    rotor_diameter_ft = EXCLUDED.rotor_diameter_ft,
    srs = EXCLUDED.srs,
    lahso = EXCLUDED.lahso,
    faa_registry = EXCLUDED.faa_registry,
    registration_count = EXCLUDED.registration_count,
    tmfs_operations_fy24 = EXCLUDED.tmfs_operations_fy24,
    remarks = EXCLUDED.remarks,
    last_update = EXCLUDED.last_update,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteAllAircraftData :exec
DELETE FROM aircraft_data;