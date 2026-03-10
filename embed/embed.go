package embed

import "embed"

//go:embed docker/docker-compose.yml
var DockerCompose []byte

//go:embed docker/postgres/Dockerfile
var PostgresDockerfile []byte

//go:embed docker/surrealdb/Dockerfile
var SurrealDBDockerfile []byte

//go:embed seed/surrealdb/00-schema.surql
var SurrealDBSchema []byte

//go:embed seed/surrealdb/01-data.surql
var SurrealDBData []byte

//go:embed seed/postgres/00-extensions.sql
var PostgresExtensions []byte

//go:embed seed/postgres/01-schema.sql
var PostgresSchema []byte

//go:embed seed/postgres/02-data.sql
var PostgresData []byte

//go:embed skills/db.md
var SkillDB []byte

//go:embed docs/feature-matrix.md
var DocsFeatureMatrix []byte

//go:embed docs/surrealdb-guide.md
var DocsSurrealDBGuide []byte

//go:embed docs/postgres-guide.md
var DocsPostgresGuide []byte

//go:embed all:docker
var DockerFS embed.FS

//go:embed all:seed
var SeedFS embed.FS

//go:embed all:skills
var SkillsFS embed.FS

//go:embed all:docs
var DocsFS embed.FS
