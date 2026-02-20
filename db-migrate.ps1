# db-migrate.ps1
param(
    [Parameter(Position=0)]
    [string]$Action = "help",
    
    [Parameter(Position=1)]
    [string]$Name = ""
)

$DB_URL = "postgresql://admin:admin@localhost:5432/auth_db?sslmode=disable"
$MIGRATIONS_PATH = "migrations"

switch ($Action) {
    "up" {
        Write-Host "🚀 Ejecutando migraciones UP..." -ForegroundColor Green
        migrate -path $MIGRATIONS_PATH -database $DB_URL up
    }
    
    "down" {
        Write-Host "⬇️  Revirtiendo última migración..." -ForegroundColor Yellow
        migrate -path $MIGRATIONS_PATH -database $DB_URL down 1
    }
    
    "down-all" {
        Write-Host "⚠️  ADVERTENCIA: Esto revertirá TODAS las migraciones" -ForegroundColor Red
        $confirm = Read-Host "¿Estás seguro? (y/N)"
        if ($confirm -eq "y" -or $confirm -eq "Y") {
            migrate -path $MIGRATIONS_PATH -database $DB_URL down -all
        }
    }
    
    "version" {
        Write-Host "📊 Versión actual:" -ForegroundColor Cyan
        migrate -path $MIGRATIONS_PATH -database $DB_URL version
    }
    
    "create" {
        if ($Name -eq "") {
            $Name = Read-Host "Nombre de la migración"
        }
        Write-Host "📝 Creando migración: $Name" -ForegroundColor Green
        migrate create -ext sql -dir $MIGRATIONS_PATH -seq $Name
    }
    
    "force" {
        if ($Name -eq "") {
            Write-Host "❌ Debes especificar una versión" -ForegroundColor Red
            exit 1
        }
        migrate -path $MIGRATIONS_PATH -database $DB_URL force $Name
    }
    
    default {
        Write-Host ""
        Write-Host "📖 Comandos disponibles:" -ForegroundColor Yellow
        Write-Host "  .\db-migrate.ps1 up          - Ejecuta migraciones" -ForegroundColor Cyan
        Write-Host "  .\db-migrate.ps1 down        - Revierte última migración" -ForegroundColor Cyan
        Write-Host "  .\db-migrate.ps1 version     - Muestra versión actual" -ForegroundColor Cyan
        Write-Host "  .\db-migrate.ps1 create X    - Crea nueva migración" -ForegroundColor Cyan
        Write-Host ""
    }
}