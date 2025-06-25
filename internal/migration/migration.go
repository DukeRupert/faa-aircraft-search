package migration

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AircraftData struct {
	ICAOCode                           string
	FAADesignator                      string
	Manufacturer                       string
	ModelFAA                           string
	ModelBADA                          string
	PhysicalClassEngine                string
	NumEngines                         *int
	AAC                                string
	AACMinimum                         string
	AACMaximum                         string
	ADG                                string
	TDG                                string
	ApproachSpeedKnot                  *int
	ApproachSpeedMinimumKnot           *int
	ApproachSpeedMaximumKnot           *int
	WingspanFtWithoutWingletsSharklets *float64
	WingspanFtWithWingletsSharklets    *float64
	LengthFt                           *float64
	TailHeightAtOEWFt                  *float64
	WheelbaseFt                        *float64
	CockpitToMainGearFt                *float64
	MainGearWidthFt                    *float64
	MTOWLB                             *int
	MALWLB                             *int
	MainGearConfig                     string
	ICAOWTC                            string
	ParkingAreaFt2                     *float64
	Class                              string
	FAAWeight                          string
	CWT                                string
	OneHalfWakeCategory                string
	TwoWakeCategoryAppxA               string
	TwoWakeCategoryAppxB               string
	RotorDiameterFt                    *float64
	SRS                                string
	LAHSO                              string
	FAARegistry                        string
	RegistrationCount                  *int
	TMFSOperationsFY24                 *int
	Remarks                            string
	LastUpdate                         string
}

// MigrateFromExcel imports aircraft data from an Excel file into the database
func MigrateFromExcel(ctx context.Context, pool *pgxpool.Pool, filePath string) error {
	log.Printf("Starting migration from Excel file: %s", filePath)

	// Open Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get all rows from ACD_Data sheet
	rows, err := f.GetRows("ACD_Data")
	if err != nil {
		return fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("no rows found in Excel file")
	}

	log.Printf("Found %d total rows (including header)", len(rows))

	// Skip header row and process data rows
	successCount := 0
	errorCount := 0
	
	for i, row := range rows[1:] {
		aircraft := parseRow(row)
		
		err := insertAircraftData(ctx, pool, aircraft)
		if err != nil {
			log.Printf("Failed to insert row %d: %v", i+2, err)
			errorCount++
			continue
		}
		
		successCount++
		if successCount%100 == 0 {
			log.Printf("Successfully processed %d rows", successCount)
		}
	}

	log.Printf("Migration completed. Success: %d, Errors: %d, Total: %d", 
		successCount, errorCount, len(rows)-1)

	return nil
}

// ClearData removes all aircraft data from the database
func ClearData(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Clearing all aircraft data...")
	
	_, err := pool.Exec(ctx, "DELETE FROM aircraft_data")
	if err != nil {
		return fmt.Errorf("failed to clear aircraft data: %w", err)
	}
	
	log.Println("Aircraft data cleared successfully")
	return nil
}

// GetRecordCount returns the number of records in the aircraft_data table
func GetRecordCount(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM aircraft_data").Scan(&count)
	return count, err
}

func parseRow(row []string) AircraftData {
	// Helper function to safely get string value
	getString := func(index int) string {
		if index < len(row) {
			return strings.TrimSpace(row[index])
		}
		return ""
	}

	// Helper function to safely parse int
	getInt := func(index int) *int {
		if index < len(row) && strings.TrimSpace(row[index]) != "" && strings.TrimSpace(row[index]) != "N/A" {
			// Remove commas for numbers like "50,045"
			cleanStr := strings.ReplaceAll(strings.TrimSpace(row[index]), ",", "")
			if val, err := strconv.Atoi(cleanStr); err == nil {
				return &val
			}
		}
		return nil
	}

	// Helper function to safely parse float
	getFloat := func(index int) *float64 {
		if index < len(row) && strings.TrimSpace(row[index]) != "" && strings.TrimSpace(row[index]) != "N/A" {
			// Remove commas for numbers
			cleanStr := strings.ReplaceAll(strings.TrimSpace(row[index]), ",", "")
			if val, err := strconv.ParseFloat(cleanStr, 64); err == nil {
				return &val
			}
		}
		return nil
	}

	return AircraftData{
		ICAOCode:                           getString(0),
		FAADesignator:                      getString(1),
		Manufacturer:                       getString(2),
		ModelFAA:                           getString(3),
		ModelBADA:                          getString(4),
		PhysicalClassEngine:                getString(5),
		NumEngines:                         getInt(6),
		AAC:                                getString(7),
		AACMinimum:                         getString(8),
		AACMaximum:                         getString(9),
		ADG:                                getString(10),
		TDG:                                getString(11),
		ApproachSpeedKnot:                  getInt(12),
		ApproachSpeedMinimumKnot:           getInt(13),
		ApproachSpeedMaximumKnot:           getInt(14),
		WingspanFtWithoutWingletsSharklets: getFloat(15),
		WingspanFtWithWingletsSharklets:    getFloat(16),
		LengthFt:                           getFloat(17),
		TailHeightAtOEWFt:                  getFloat(18),
		WheelbaseFt:                        getFloat(19),
		CockpitToMainGearFt:                getFloat(20),
		MainGearWidthFt:                    getFloat(21),
		MTOWLB:                             getInt(22),
		MALWLB:                             getInt(23),
		MainGearConfig:                     getString(24),
		ICAOWTC:                            getString(25),
		ParkingAreaFt2:                     getFloat(26),
		Class:                              getString(27),
		FAAWeight:                          getString(28),
		CWT:                                getString(29),
		OneHalfWakeCategory:                getString(30),
		TwoWakeCategoryAppxA:               getString(31),
		TwoWakeCategoryAppxB:               getString(32),
		RotorDiameterFt:                    getFloat(33),
		SRS:                                getString(34),
		LAHSO:                              getString(35),
		FAARegistry:                        getString(36),
		RegistrationCount:                  getInt(37),
		TMFSOperationsFY24:                 getInt(38),
		Remarks:                            getString(39),
		LastUpdate:                         getString(40),
	}
}

func insertAircraftData(ctx context.Context, pool *pgxpool.Pool, aircraft AircraftData) error {
	query := `
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
	`

	// Use a context with timeout for the query
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := pool.Exec(queryCtx, query,
		aircraft.ICAOCode, aircraft.FAADesignator, aircraft.Manufacturer, aircraft.ModelFAA, aircraft.ModelBADA,
		aircraft.PhysicalClassEngine, aircraft.NumEngines, aircraft.AAC, aircraft.AACMinimum, aircraft.AACMaximum,
		aircraft.ADG, aircraft.TDG, aircraft.ApproachSpeedKnot, aircraft.ApproachSpeedMinimumKnot, aircraft.ApproachSpeedMaximumKnot,
		aircraft.WingspanFtWithoutWingletsSharklets, aircraft.WingspanFtWithWingletsSharklets,
		aircraft.LengthFt, aircraft.TailHeightAtOEWFt, aircraft.WheelbaseFt, aircraft.CockpitToMainGearFt,
		aircraft.MainGearWidthFt, aircraft.MTOWLB, aircraft.MALWLB, aircraft.MainGearConfig, aircraft.ICAOWTC,
		aircraft.ParkingAreaFt2, aircraft.Class, aircraft.FAAWeight, aircraft.CWT, aircraft.OneHalfWakeCategory,
		aircraft.TwoWakeCategoryAppxA, aircraft.TwoWakeCategoryAppxB, aircraft.RotorDiameterFt,
		aircraft.SRS, aircraft.LAHSO, aircraft.FAARegistry, aircraft.RegistrationCount, aircraft.TMFSOperationsFY24,
		aircraft.Remarks, aircraft.LastUpdate,
	)

	return err
}