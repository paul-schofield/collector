// Scrape `information_schema.table_statistics`.

package collector

import (
	"context"
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const schemaStatQuery = `
		SELECT 
			TABLE_SCHEMA, 
			SUM(ROWS_READ) AS ROWS_READ, 
			SUM(ROWS_CHANGED) AS ROWS_CHANGED, 
			SUM(ROWS_CHANGED_X_INDEXES) AS ROWS_CHANGED_X_INDEXES 
		FROM information_schema.TABLE_STATISTICS 
		GROUP BY TABLE_SCHEMA;
		`

// Metric descriptors.
var (
	infoSchemaStatsRowsReadDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "schema_statistics_rows_read_total"),
		"The number of rows read from the schema.",
		[]string{"schema"}, nil,
	)
	infoSchemaStatsRowsChangedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "schema_statistics_rows_changed_total"),
		"The number of rows changed in the schema.",
		[]string{"schema"}, nil,
	)
	infoSchemaStatsRowsChangedXIndexesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "schema_statistics_rows_changed_x_indexes_total"),
		"The number of rows changed in the schema, multiplied by the number of indexes changed.",
		[]string{"schema"}, nil,
	)
)

// ScrapeSchemaStat collects from `information_schema.table_statistics` grouped by schema.
type ScrapeSchemaStat struct{}

// Name of the Scraper. Should be unique.
func (ScrapeSchemaStat) Name() string {
	return "info_schema.schemastats"
}

// Help describes the role of the Scraper.
func (ScrapeSchemaStat) Help() string {
	return "If running with userstat=1, set to true to collect schema statistics"
}

// Version of MySQL from which scraper is available.
func (ScrapeSchemaStat) Version() float64 {
	return 5.1
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeSchemaStat) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	var varName, varVal string

	err := db.QueryRowContext(ctx, userstatCheckQuery).Scan(&varName, &varVal)
	if err != nil {
		log.Debugln("Detailed schema stats are not available.")
		return nil
	}
	if varVal == "OFF" {
		log.Debugf("MySQL @@%s is OFF.", varName)
		return nil
	}

	informationSchemaTableStatisticsRows, err := db.QueryContext(ctx, schemaStatQuery)
	if err != nil {
		return err
	}
	defer informationSchemaTableStatisticsRows.Close()

	var (
		tableSchema         string
		rowsRead            uint64
		rowsChanged         uint64
		rowsChangedXIndexes uint64
	)

	for informationSchemaTableStatisticsRows.Next() {
		err = informationSchemaTableStatisticsRows.Scan(
			&tableSchema,
			&rowsRead,
			&rowsChanged,
			&rowsChangedXIndexes,
		)

		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			infoSchemaStatsRowsReadDesc, prometheus.CounterValue, float64(rowsRead),
			tableSchema,
		)
		ch <- prometheus.MustNewConstMetric(
			infoSchemaStatsRowsChangedDesc, prometheus.CounterValue, float64(rowsChanged),
			tableSchema,
		)
		ch <- prometheus.MustNewConstMetric(
			infoSchemaStatsRowsChangedXIndexesDesc, prometheus.CounterValue, float64(rowsChangedXIndexes),
			tableSchema,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeSchemaStat{}
