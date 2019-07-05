package cmd

import (
	"database/sql"
	"strings"
	"time"

	"github.com/buxtronix/b6max"
	_ "github.com/mattn/go-sqlite3"
)

const (
	createStmt = `
CREATE TABLE IF NOT EXISTS watchLog (
	programId INTEGER NOT NULL,
	timestamp INTEGER NOT NULL,
	elapsed INTEGER NOT NULL,
	tags TEXT,
	state text,
	errorCode INTEGER,
	timer INTEGER,
	milliAmpHour INTEGER,
	milliAmps INTEGER,
	milliVolts INTEGER,
	temperatureExternal INTEGER,
	temperatureInternal INTEGER,
	impedance INTEGER,
	cell1MilliVolts INTEGER,
	cell2MilliVolts INTEGER,
	cell3MilliVolts INTEGER,
	cell4MilliVolts INTEGER,
	cell5MilliVolts INTEGER,
	cell6MilliVolts INTEGER,
	cell7MilliVolts INTEGER,
	cell8MilliVolts INTEGER
);
CREATE TABLE IF NOT EXISTS programLog (
	programId INTEGER NOT NULL,
	timestamp INTEGER NOT NULL,
	batteryType TEXT,
	tags TEXT,
	cells INTEGER,
	programMode TEXT,
	chargeCurrent INTEGER,
	dischargeCurrent INTEGER,
	dischargeCutoff INTEGER,
	chargeCutoff INTEGER,
	repeakCount INTEGER,
	cycleType INTEGER,
	cycleCount INTEGER,
	trickleCurrent INTEGER
);
`
)

type database struct {
	db        *sql.DB
	programId int64
	tags      []string
}

func (d *database) Open(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	d.db = db

	if _, err = d.db.Exec(createStmt); err != nil {
		return err
	}
	return nil
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) GetHighestProgramId() (int64, error) {
	rows, err := d.db.Query("select max(programId) from programLog")
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var programId sql.NullInt64
	for rows.Next() {
		if err = rows.Scan(&programId); err != nil {
			return 0, err
		}
	}
	return programId.Int64, nil
}

func (d *database) setNextProgramId() error {
	pgmId, err := d.GetHighestProgramId()
	if err != nil {
		return err
	}
	d.programId = pgmId + 1
	return nil
}

func (d *database) writeProgram(pgm *b6max.ProgramStart, tags []string) error {
	if d.programId == 0 {
		if err := d.setNextProgramId(); err != nil {
			return err
		}
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO programLog (
		programId, timestamp, batteryType, tags, cells, programMode,
		chargeCurrent, dischargeCurrent, dischargeCutoff, chargeCutoff,
		repeakCount, cycleType, cycleCount, trickleCurrent) values (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(
		d.programId, time.Now().Unix(), pgm.BatteryType.String(),
		strings.Join(tags, ","), pgm.Cells, pgm.PwmMode.String(),
		pgm.ChargeCurrent, pgm.DischargeCurrent, pgm.DischargeCutoff, pgm.ChargeCutoff,
		pgm.RePeakCycleInfo, pgm.RePeakCycleInfo, pgm.CycleCount,
		pgm.Trickle); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *database) writeInfo(info *b6max.ProgramState, tags []string) error {
	if d.programId == 0 {
		if err := d.setNextProgramId(); err != nil {
			return err
		}
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO watchLog (
		programId, timestamp, elapsed, tags, state, errorCode, timer,
		milliAmpHour, milliAmps,milliVolts, temperatureExternal,
		temperatureInternal, impedance, cell1Millivolts, cell2MilliVolts,
		cell3MilliVolts, cell4MilliVolts, cell5MilliVolts, cell6MilliVolts)
		values
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()
	if _, err = stmt.Exec(
		d.programId, now, now-d.programId, strings.Join(tags, ","), info.WorkState.String(),
		info.ErrorCodeMah, info.Time, info.ErrorCodeMah, info.MilliAmp, info.MilliVolt,
		info.TemperatureExternal, info.TemperatureInternal, info.Impedance,
		info.Cells[0], info.Cells[1], info.Cells[2], info.Cells[3], info.Cells[4],
		info.Cells[5]); err != nil {
		return err
	}
	return tx.Commit()
}
