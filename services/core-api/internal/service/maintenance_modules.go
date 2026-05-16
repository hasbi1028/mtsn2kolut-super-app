package service

import "strings"

const (
	MaintenanceModeOff      = "off"
	MaintenanceModeGlobal   = "global"
	MaintenanceModeModule   = "module"
	MaintenanceModeReadOnly = "read_only"

	MaintenanceModuleGlobal        = "global"
	MaintenanceModuleAuth          = "auth"
	MaintenanceModuleDashboard     = "dashboard"
	MaintenanceModuleAkademik      = "akademik"
	MaintenanceModuleStudents      = "students"
	MaintenanceModuleBankSoal      = "bank_soal"
	MaintenanceModuleCBT           = "cbt"
	MaintenanceModulePusaka        = "pusaka"
	MaintenanceModuleBackupRestore = "backup_restore"
	MaintenanceModuleSettings      = "settings"
)

var validMaintenanceModules = map[string]struct{}{
	MaintenanceModuleGlobal:        {},
	MaintenanceModuleAuth:          {},
	MaintenanceModuleDashboard:     {},
	MaintenanceModuleAkademik:      {},
	MaintenanceModuleStudents:      {},
	MaintenanceModuleBankSoal:      {},
	MaintenanceModuleCBT:           {},
	MaintenanceModulePusaka:        {},
	MaintenanceModuleBackupRestore: {},
	MaintenanceModuleSettings:      {},
}

var maintenanceAPIPrefixes = map[string][]string{
	MaintenanceModuleAuth: {
		"/api/auth",
	},
	MaintenanceModuleDashboard: {
		"/api/internal-analytics",
		"/api/notifications",
	},
	MaintenanceModuleAkademik: {
		"/api/academic",
		"/api/journal",
		"/api/grades",
	},
	MaintenanceModuleStudents: {
		"/api/students",
		"/api/kesiswaan",
		"/api/parents",
		"/api/portal",
	},
	MaintenanceModuleBankSoal: {
		"/api/bank-soal",
		"/api/cbt/questions",
		"/api/cbt/assets",
	},
	MaintenanceModuleCBT: {
		"/api/asesmen",
		"/api/cbt/packages",
		"/api/cbt/events",
		"/api/cbt/sessions",
		"/api/cbt/non-test-assessments",
		"/api/exam",
	},
	MaintenanceModulePusaka: {
		"/api/pusaka",
	},
	MaintenanceModuleBackupRestore: {
		"/api/system/backups",
	},
	MaintenanceModuleSettings: {
		"/api/settings",
		"/api/users",
		"/api/rbac",
		"/api/school-profile",
		"/api/branding",
		"/api/system/maintenance",
	},
}

func MaintenanceModules() []string {
	return []string{
		MaintenanceModuleGlobal,
		MaintenanceModuleAuth,
		MaintenanceModuleDashboard,
		MaintenanceModuleAkademik,
		MaintenanceModuleStudents,
		MaintenanceModuleBankSoal,
		MaintenanceModuleCBT,
		MaintenanceModulePusaka,
		MaintenanceModuleBackupRestore,
		MaintenanceModuleSettings,
	}
}

func IsValidMaintenanceModule(module string) bool {
	_, ok := validMaintenanceModules[strings.TrimSpace(strings.ToLower(module))]
	return ok
}

func MaintenanceModuleMatchesAPIPath(module, path string) bool {
	module = strings.TrimSpace(strings.ToLower(module))
	if module == MaintenanceModuleGlobal {
		return strings.HasPrefix(path, "/api/")
	}
	for _, prefix := range maintenanceAPIPrefixes[module] {
		if matchesMaintenancePathSegment(path, prefix) {
			return true
		}
	}
	return false
}

func MaintenanceModulesMatchAPIPath(modules []string, path string) bool {
	for _, module := range modules {
		if MaintenanceModuleMatchesAPIPath(module, path) {
			return true
		}
	}
	return false
}

func matchesMaintenancePathSegment(path, prefix string) bool {
	prefix = strings.TrimRight(prefix, "/")
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
