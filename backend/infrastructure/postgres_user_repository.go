package infrastructure

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ileego/go_browser_capture/backend/domain"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	var createdAt, updatedAt time.Time
	var roleID, roleName, roleDescription string
	var rolePermissions []string

	query := `SELECT u.id, u.username, u.password, u.email, u.role_id, u.is_active, u.created_at, u.updated_at,
	                 r.id as role_id, r.name as role_name, r.description as role_description, r.permissions as role_permissions
			  FROM users u
			  LEFT JOIN roles r ON u.role_id = r.id
			  WHERE u.id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email, &user.RoleID, &user.IsActive,
		&createdAt, &updatedAt,
		&roleID, &roleName, &roleDescription, &rolePermissions,
	)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	if roleID != "" {
		user.Role = &domain.Role{
			ID:          roleID,
			Name:        roleName,
			Description: roleDescription,
			Permissions: rolePermissions,
		}
	}

	return &user, nil
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	var createdAt, updatedAt time.Time
	var roleID, roleName, roleDescription string
	var rolePermissions []string

	query := `SELECT u.id, u.username, u.password, u.email, u.role_id, u.is_active, u.created_at, u.updated_at,
	                 r.id as role_id, r.name as role_name, r.description as role_description, r.permissions as role_permissions
			  FROM users u
			  LEFT JOIN roles r ON u.role_id = r.id
			  WHERE u.username = $1`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email, &user.RoleID, &user.IsActive,
		&createdAt, &updatedAt,
		&roleID, &roleName, &roleDescription, &rolePermissions,
	)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	if roleID != "" {
		user.Role = &domain.Role{
			ID:          roleID,
			Name:        roleName,
			Description: roleDescription,
			Permissions: rolePermissions,
		}
	}

	return &user, nil
}

func (r *PostgresUserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT u.id, u.username, u.password, u.email, u.role_id, u.is_active, u.created_at, u.updated_at,
	                 r.id as role_id, r.name as role_name, r.description as role_description, r.permissions as role_permissions
			  FROM users u
			  LEFT JOIN roles r ON u.role_id = r.id
			  ORDER BY u.created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var createdAt, updatedAt time.Time
		var roleID, roleName, roleDescription string
		var rolePermissions []string

		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Email, &user.RoleID, &user.IsActive,
			&createdAt, &updatedAt,
			&roleID, &roleName, &roleDescription, &rolePermissions,
		)
		if err != nil {
			return nil, err
		}

		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt

		if roleID != "" {
			user.Role = &domain.Role{
				ID:          roleID,
				Name:        roleName,
				Description: roleDescription,
				Permissions: rolePermissions,
			}
		}

		users = append(users, &user)
	}

	return users, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, username, password, email, role_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Password, user.Email,
		user.RoleID, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET username = $1, email = $2, role_id = $3, is_active = $4, updated_at = $5 WHERE id = $6`

	_, err := r.db.Exec(ctx, query,
		user.Username, user.Email, user.RoleID, user.IsActive, time.Now(), user.ID,
	)
	return err
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

type PostgresRoleRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRoleRepository(db *pgxpool.Pool) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) FindByID(ctx context.Context, id string) (*domain.Role, error) {
	var role domain.Role
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, `SELECT id, name, description, permissions, created_at, updated_at FROM roles WHERE id = $1`, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.Permissions, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	role.CreatedAt = createdAt
	role.UpdatedAt = updatedAt

	return &role, nil
}

func (r *PostgresRoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, `SELECT id, name, description, permissions, created_at, updated_at FROM roles WHERE name = $1`, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.Permissions, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	role.CreatedAt = createdAt
	role.UpdatedAt = updatedAt

	return &role, nil
}

func (r *PostgresRoleRepository) FindAll(ctx context.Context) ([]*domain.Role, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, description, permissions, created_at, updated_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		var role domain.Role
		var createdAt, updatedAt time.Time

		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.Permissions, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		role.CreatedAt = createdAt
		role.UpdatedAt = updatedAt
		roles = append(roles, &role)
	}

	return roles, nil
}

func (r *PostgresRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	query := `INSERT INTO roles (id, name, description, permissions, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query, role.ID, role.Name, role.Description, role.Permissions, role.CreatedAt, role.UpdatedAt)
	return err
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	query := `UPDATE roles SET name = $1, description = $2, permissions = $3, updated_at = $4 WHERE id = $5`

	_, err := r.db.Exec(ctx, query, role.Name, role.Description, role.Permissions, time.Now(), role.ID)
	return err
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM roles WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	return err
}
