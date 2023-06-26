package db

/*
All the queries, variables and statuses the program queries for
Having them here declutters the program
*/

func ProcesslistLongQuery() string {
	return `SELECT pps.PROCESSLIST_COMMAND AS command,
				pps.THREAD_ID AS thd_id, pps.PROCESSLIST_ID AS conn_id,
			conattr_pid.ATTR_VALUE AS pid, pps.PROCESSLIST_STATE AS state,
			if((pps.NAME in ('thread/sql/one_connection','thread/thread_pool/tp_one_connection')),
			concat(pps.PROCESSLIST_USER,'@',pps.PROCESSLIST_HOST),
			replace(pps.NAME,'thread/','')) AS user,
			pps.PROCESSLIST_DB AS db, 
			IF(CHAR_LENGTH(pps.PROCESSLIST_INFO) > 64, REPLACE(CONCAT(LEFT(pps.PROCESSLIST_INFO, 30), ' ... ', RIGHT(pps.PROCESSLIST_INFO, 30)), '\n', ' '), REPLACE(pps.PROCESSLIST_INFO, '\n', ' ')) AS current_statement,
			if(isnull(esc.END_EVENT_ID), esc.TIMER_WAIT,NULL) AS statement_latency,
			esc.LOCK_TIME AS lock_latency,
			if(isnull(esc.END_EVENT_ID),esc.TIMER_WAIT,0) AS sort_time
			from (performance_schema.threads pps
			left join performance_schema.events_statements_current esc
			on (pps.THREAD_ID = esc.THREAD_ID))
									left join performance_schema.session_connect_attrs conattr_pid
									on((conattr_pid.PROCESSLIST_ID = pps.PROCESSLIST_ID) and (conattr_pid.ATTR_NAME = '_pid'))
			where pps.PROCESSLIST_ID is not null
			and pps.PROCESSLIST_COMMAND <> 'Daemon'
			`
}

func InnoDBLongParams() []string {
	return []string{"innodb_buffer_pool_read_requests%", "innodb_buffer_pool_write_requests%",
		"innodb_buffer_pool_pages_dirty", "innodb_buffer_pool_reads", "innodb_buffer_pool_writes",
		"innodb_os_log_pending_writes", "handler_read_first", "handler_read_key", "handler_read_next",
		"handler_read_prev", "handler_read_rnd", "handler_read_rnd_next", "innodb_data_pending_fsyncs",
		"innodb_os_log_pending_fsyncs"}
}

func SelectLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'SELECT%', count_star, 0)) AS select_count
			FROM performance_schema.events_statements_summary_by_digest;`
}

func InsertsLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'INSERTST%', count_star, 0)) AS insert_count
			FROM performance_schema.events_statements_summary_by_digest;`
}

func UpdatesLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'UPDATES%', count_star, 0)) AS update_count
			FROM performance_schema.events_statements_summary_by_digest;`
}

func DeletesLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'DELETES%', count_star, 0)) AS delete_count
			FROM performance_schema.events_statements_summary_by_digest;`
}

func UserMemoryShortQuery() string {
	return `SELECT user, current_allocated, current_max_alloc
			FROM sys.memory_by_user_by_current_bytes
			WHERE user != "background";`
}

func GlobalAllocatedShortQuery() string {
	return `SELECT total_allocated FROM sys.memory_global_total;`
}

func SpecificAllocatedLongQuery() string {
	return `SELECT SUBSTRING_INDEX(event_name,'/',2) AS code_area,
			format_bytes(SUM(current_alloc)) AS current_alloc,
			sum(current_alloc) current_alloc_num
			FROM sys.x$memory_global_by_current_bytes
			GROUP BY SUBSTRING_INDEX(event_name,'/',2)
			ORDER BY SUM(current_alloc) DESC;`
}

func RamNDiskLongQuery() string {
	return `SELECT event_name,
			format_bytes(CURRENT_NUMBER_OF_BYTES_USED) AS current_alloc,
			format_bytes(HIGH_NUMBER_OF_BYTES_USED) AS high_alloc
			FROM performance_schema.memory_summary_global_by_event_name
			WHERE event_name LIKE 'memory/temptable/%';`
}

func CheckpointInfoLongQuery() string {
	return `SELECT CONCAT(
			(
			SELECT FORMAT_BYTES(
			STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
			)
			FROM performance_schema.log_status),
			" / ",
			format_bytes(
			(SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_file_size'
			)  * (
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_files_in_group'))
			) CheckpointInfo;`
}

func CheckpointAgePctLongQuery() string {
	return `SELECT ROUND(((
			SELECT STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
			FROM performance_schema.log_status) / ((
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_file_size'
			) * (
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_files_in_group')) * 100));`
}

func ErrorLogShortQuery() string {
	return `SELECT *, cast(unix_timestamp(logged)*1000000 as unsigned) logged_int FROM performance_schema.error_log`
}

func LocksLongQuery() string {
	return `SELECT
			r.trx_id waiting_trx_id,
			r.trx_mysql_thread_id waiting_thread,
			r.trx_query waiting_query,
			b.trx_id blocking_trx_id,
			b.trx_mysql_thread_id blocking_thread,
			b.trx_query blocking_query
			FROM       performance_schema.data_lock_waits w
			INNER JOIN information_schema.innodb_trx b
			ON b.trx_id = w.blocking_engine_transaction_id
			INNER JOIN information_schema.innodb_trx r
			ON r.trx_id = w.requesting_engine_transaction_id;`
}
