@echo off
SET DB_URL=postgres://postgres:postgres@localhost:5432/goshort?sslmode=disable


REM Cek argumen pertama: up, down, create, version, force, dll
IF "%1"=="up" (
    migrate -database %DB_URL% -path db/migrations up
) ELSE IF "%1"=="down" (
    migrate -database %DB_URL% -path db/migrations down
) ELSE IF "%1"=="force" (
    migrate -database %DB_URL% -path db/migrations force %2
) ELSE IF "%1"=="version" (
    migrate -database %DB_URL% -path db/migrations version
) ELSE IF "%1"=="drop" (
    migrate -database %DB_URL% -path db/migrations drop
) ELSE IF "%1"=="create" (
    IF "%2"=="" (
        echo Please provide a migration name.
    ) ELSE (
        migrate create -ext sql -dir db/migrations -seq %2
    )
) ELSE (
    echo Usage:
    echo   migrate.bat up
    echo   migrate.bat down
    echo   migrate.bat force VERSION
    echo   migrate.bat version
    echo   migrate.bat drop
    echo   migrate.bat create MIGRATION_NAME
)
