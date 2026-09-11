package service

import (
	"context"

	"ops-dump/internal/dao"
	"ops-dump/internal/model/entity"
)

// builtinTemplates are seeded on first start with enabled=0.
// Operators enable them from the page; connection and output settings come from .env.
var builtinTemplates = []entity.Job{
	{
		Name:          "mysql-full",
		Kind:          "mysql",
		Enabled:       0,
		Cron:          "0 0 2 * * *",
		Description:   "MySQL 通过 mysqldump 全量备份，连接和输出设置从环境变量读取。",
		Command:       "BACKUP_DATE=$(date +%Y%m%d)\nBACKUP_FILE=${BACKUP_DIR}/mysql_full_${BACKUP_DATE}.sql.gz\n\nmysqldump --single-transaction --quick --routines --triggers --events \\\n  --set-gtid-purged=OFF --default-character-set=utf8mb4 \\\n  -h${MYSQL_HOST:-127.0.0.1} -P${MYSQL_PORT:-3306} \\\n  -uroot -p${MYSQL_ROOT_PASSWORD} \\\n  --databases ${MYSQL_DATABASE} | gzip -9 > ${BACKUP_FILE}",
		VerifyCommand: "BACKUP_FILE=$(ls -t ${BACKUP_DIR}/mysql_full_*.sql.gz 2>/dev/null | head -1)\ntest -s ${BACKUP_FILE}",
	},
	{
		Name:          "tdengine-full",
		Kind:          "tdengine",
		Enabled:       0,
		Cron:          "0 0 3 * * *",
		Description:   "TDengine 通过 taosdump 全量备份，连接和输出设置从环境变量读取。",
		Command:       "BACKUP_DATE=$(date +%Y%m%d)\nTEMP_DIR=/tmp/taos_backup_${BACKUP_DATE}\nTARGET_DIR=${BACKUP_DIR}/${TDENGINE_DATABASE}_${BACKUP_DATE}\n\nmkdir -p ${BACKUP_DIR}\ntaosdump -h ${TDENGINE_HOST:-127.0.0.1} -P ${TDENGINE_PORT:-6030} \\\n  -uroot -p${TDENGINE_ROOT_PASSWORD} \\\n  -o ${TEMP_DIR} -T 4 -D ${TDENGINE_DATABASE}\ncp -r ${TEMP_DIR} ${TARGET_DIR}\nrm -rf ${TEMP_DIR}",
		VerifyCommand: "BACKUP_DATE=$(date +%Y%m%d)\ntest -n \"$(find ${BACKUP_DIR}/${TDENGINE_DATABASE}_${BACKUP_DATE} -type f 2>/dev/null | head -1)\"",
	},
	{
		Name:          "minio-mirror",
		Kind:          "minio",
		Enabled:       0,
		Cron:          "0 0 4 * * *",
		Description:   "MinIO 通过本地 mc 镜像 Bucket 到宿主机目录，连接和输出设置从环境变量读取（要求宿主机已安装 mc）。",
		Command:       "mc alias set dst ${MINIO_ENDPOINT} \"${MINIO_ROOT_USER}\" \"${MINIO_ROOT_PASSWORD}\" 2>/dev/null\nmc mirror --overwrite dst/${MINIO_BUCKET} ${BACKUP_DIR}/${MINIO_BUCKET}",
		VerifyCommand: "test -n \"$(find ${BACKUP_DIR}/${MINIO_BUCKET} -type f 2>/dev/null | head -1)\"",
	},
}

// TemplateService seeds the built-in backup templates.
type TemplateService struct{}

// NewTemplate creates a TemplateService instance.
func NewTemplate() *TemplateService { return &TemplateService{} }

// SeedTemplates inserts the built-in templates only if their names do not exist yet.
func (s *TemplateService) SeedTemplates(ctx context.Context) error {
	for i := range builtinTemplates {
		tpl := builtinTemplates[i]
		n, err := dao.Jobs.Ctx(ctx).Where("name", tpl.Name).Count()
		if err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		tpl.CreatedAt = now()
		tpl.UpdatedAt = now()
		if _, err := dao.Jobs.Ctx(ctx).Insert(dataNoId(tpl)); err != nil {
			return err
		}
	}
	return nil
}
