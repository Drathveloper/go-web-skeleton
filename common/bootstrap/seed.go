package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	commonconfig "github.com/Drathveloper/go-web-skeleton/common/config"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
	securitydomain "github.com/Drathveloper/go-web-skeleton/security/domain"
)

const seedAdminErrMsg = "seed administrator failed"

// seedAdministrator creates the first administrator when the database has none.
//
// Every route that manages users is behind Authorize(AdminRole), so without
// this a freshly created project has no way to log in at all: there is no admin
// to create the first admin. It runs only on an empty user table, takes the
// credentials from the environment, and refuses to invent a default — a
// password baked into a template is a password shipped to every project
// generated from it.
func seedAdministrator(ctx context.Context, container *wire.Container) error {
	users, err := container.UserManagementService.ListUsers(ctx, &commondomain.Pagination{Page: 1, Size: 1})
	if err != nil {
		return fmt.Errorf("%s: %w", seedAdminErrMsg, err)
	}
	if len(users) > 0 {
		return nil
	}

	username := commonconfig.Env.SeedAdminUsername
	password := commonconfig.Env.SeedAdminPassword
	if username == "" || password == "" {
		slog.Warn("no users exist and no seed administrator configured; " +
			"set SEED_ADMIN_USERNAME and SEED_ADMIN_PASSWORD to create the first one")
		return nil
	}

	user := &securitydomain.User{
		Username: username,
		Password: password,
		Roles:    []commondomain.Role{commondomain.AdminRole},
	}
	if err = container.UserManagementService.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("%s: %w", seedAdminErrMsg, err)
	}
	slog.Info("seeded the first administrator", slog.String("username", username))

	return nil
}
