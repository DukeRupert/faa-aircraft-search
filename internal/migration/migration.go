package migration

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dukerupert/faa-aircraft-search/internal/database"
	"github.com/dukerupert/faa-aircraft-search/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xuri/excelize/v2"
)

type AircraftData struct {
	ICAOCode                           string
	FAADesignator                      string
	Manufacturer                       string
	ModelFAA                           string
	ModelBADA                          string
	PhysicalClassEngine                string
	NumEngines                         *int32
	AAC                                string
	AACMinimum                         string
	AACMaximum                         string
	ADG                                string
	TDG                                string
	ApproachSpeedKnot                  *int32
	ApproachSpeedMinimumKnot           *int32
	ApproachSpeedMaximumKnot           *int32
	WingspanFtWithoutWingletsSharklets *float64
	WingspanFtWithWingletsSharklets    *float64
	LengthFt                           *float64
	TailHeightAtOEWFt                  *float64
	WheelbaseFt                        *float64
	CockpitToMainGearFt                *float64
	MainGearWidthFt                    *float64
	MTOWLB                             *int32
	MALWLB                             *int32
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
	RegistrationCount                  *int32
	TMFSOperationsFY24                 *int32
	Remarks                            string
	LastUpdate                         string
}

// MigrateFromExcel imports aircraft data from an Excel file into the database
func MigrateFromExcel(ctx context.Context, database *database.Database, filePath string) error {
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
		
		err := insertAircraftData(ctx, database, aircraft)
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
func ClearData(ctx context.Context, database *database.Database) error {
	log.Println("Clearing all aircraft data...")
	
	err := database.Queries.DeleteAllAircraftData(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear aircraft data: %w", err)
	}
	
	log.Println("Aircraft data cleared successfully")
	return nil
}

// GetRecordCount returns the number of records in the aircraft_data table
func GetRecordCount(ctx context.Context, database *database.Database) (int64, error) {
	count, err := database.Queries.CountAircraft(ctx)
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

	// Helper function to safely parse int32
	getInt32 := func(index int) *int32 {
		if index < len(row) && strings.TrimSpace(row[index]) != "" && strings.TrimSpace(row[index]) != "N/A" {
			// Remove commas for numbers like "50,045"
			cleanStr := strings.ReplaceAll(strings.TrimSpace(row[index]), ",", "")
			if val, err := strconv.ParseInt(cleanStr, 10, 32); err == nil {
				val32 := int32(val)
				return &val32
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
		NumEngines:                         getInt32(6),
		AAC:                                getString(7),
		AACMinimum:                         getString(8),
		AACMaximum:                         getString(9),
		ADG:                                getString(10),
		TDG:                                getString(11),
		ApproachSpeedKnot:                  getInt32(12),
		ApproachSpeedMinimumKnot:           getInt32(13),
		ApproachSpeedMaximumKnot:           getInt32(14),
		WingspanFtWithoutWingletsSharklets: getFloat(15),
		WingspanFtWithWingletsSharklets:    getFloat(16),
		LengthFt:                           getFloat(17),
		TailHeightAtOEWFt:                  getFloat(18),
		WheelbaseFt:                        getFloat(19),
		CockpitToMainGearFt:                getFloat(20),
		MainGearWidthFt:                    getFloat(21),
		MTOWLB:                             getInt32(22),
		MALWLB:                             getInt32(23),
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
		RegistrationCount:                  getInt32(37),
		TMFSOperationsFY24:                 getInt32(38),
		Remarks:                            getString(39),
		LastUpdate:                         getString(40),
	}
}

func insertAircraftData(ctx context.Context, database *database.Database, aircraft AircraftData) error {
	// Helper function to convert string to pgtype.Text
	stringToPgText := func(s string) pgtype.Text {
		if s == "" {
			return pgtype.Text{Valid: false}
		}
		return pgtype.Text{String: s, Valid: true}
	}

	// Helper function to convert *int32 to pgtype.Int4
	int32ToPgInt4 := func(i *int32) pgtype.Int4 {
		if i == nil {
			return pgtype.Int4{Valid: false}
		}
		return pgtype.Int4{Int32: *i, Valid: true}
	}

	// Helper function to convert *float64 to pgtype.Numeric
	float64ToPgNumeric := func(f *float64) pgtype.Numeric {
		if f == nil {
			return pgtype.Numeric{Valid: false}
		}
		var numeric pgtype.Numeric
		err := numeric.Scan(*f)
		if err != nil {
			return pgtype.Numeric{Valid: false}
		}
		return numeric
	}

	// Convert our internal struct to SQLC parameters
	params := db.UpsertAircraftDataParams{
		IcaoCode:                         stringToPgText(aircraft.ICAOCode),
		FaaDesignator:                    stringToPgText(aircraft.FAADesignator),
		Manufacturer:                     stringToPgText(aircraft.Manufacturer),
		ModelFaa:                         stringToPgText(aircraft.ModelFAA),
		ModelBada:                        stringToPgText(aircraft.ModelBADA),
		PhysicalClassEngine:              stringToPgText(aircraft.PhysicalClassEngine),
		NumEngines:                       int32ToPgInt4(aircraft.NumEngines),
		Aac:                              stringToPgText(aircraft.AAC),
		AacMinimum:                       stringToPgText(aircraft.AACMinimum),
		AacMaximum:                       stringToPgText(aircraft.AACMaximum),
		Adg:                              stringToPgText(aircraft.ADG),
		Tdg:                              stringToPgText(aircraft.TDG),
		ApproachSpeedKnot:                int32ToPgInt4(aircraft.ApproachSpeedKnot),
		ApproachSpeedMinimumKnot:         int32ToPgInt4(aircraft.ApproachSpeedMinimumKnot),
		ApproachSpeedMaximumKnot:         int32ToPgInt4(aircraft.ApproachSpeedMaximumKnot),
		WingspanFtWithoutWingletsSharklets: float64ToPgNumeric(aircraft.WingspanFtWithoutWingletsSharklets),
		WingspanFtWithWingletsSharklets:    float64ToPgNumeric(aircraft.WingspanFtWithWingletsSharklets),
		LengthFt:                           float64ToPgNumeric(aircraft.LengthFt),
		TailHeightAtOewFt:                  float64ToPgNumeric(aircraft.TailHeightAtOEWFt),
		WheelbaseFt:                        float64ToPgNumeric(aircraft.WheelbaseFt),
		CockpitToMainGearFt:                float64ToPgNumeric(aircraft.CockpitToMainGearFt),
		MainGearWidthFt:                    float64ToPgNumeric(aircraft.MainGearWidthFt),
		MtowLb:                             int32ToPgInt4(aircraft.MTOWLB),
		MalwLb:                             int32ToPgInt4(aircraft.MALWLB),
		MainGearConfig:                     stringToPgText(aircraft.MainGearConfig),
		IcaoWtc:                            stringToPgText(aircraft.ICAOWTC),
		ParkingAreaFt2:                     float64ToPgNumeric(aircraft.ParkingAreaFt2),
		Class:                              stringToPgText(aircraft.Class),
		FaaWeight:                          stringToPgText(aircraft.FAAWeight),
		Cwt:                                stringToPgText(aircraft.CWT),
		OneHalfWakeCategory:                stringToPgText(aircraft.OneHalfWakeCategory),
		TwoWakeCategoryAppxA:               stringToPgText(aircraft.TwoWakeCategoryAppxA),
		TwoWakeCategoryAppxB:               stringToPgText(aircraft.TwoWakeCategoryAppxB),
		RotorDiameterFt:                    float64ToPgNumeric(aircraft.RotorDiameterFt),
		Srs:                                stringToPgText(aircraft.SRS),
		Lahso:                              stringToPgText(aircraft.LAHSO),
		FaaRegistry:                        stringToPgText(aircraft.FAARegistry),
		RegistrationCount:                  int32ToPgInt4(aircraft.RegistrationCount),
		TmfsOperationsFy24:                 int32ToPgInt4(aircraft.TMFSOperationsFY24),
		Remarks:                            stringToPgText(aircraft.Remarks),
		LastUpdate:                         stringToPgText(aircraft.LastUpdate),
	}

	// Use SQLC-generated upsert function
	_, err := database.Queries.UpsertAircraftData(ctx, params)
	return err
}