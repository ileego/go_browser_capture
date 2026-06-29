package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameExists  = errors.New("username already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserDisabled    = errors.New("user is disabled")
	ErrRoleNotFound    = errors.New("role not found")
)

type UserService struct {
	userRepo UserRepository
	roleRepo RoleRepository
}

func NewUserService(userRepo UserRepository, roleRepo RoleRepository) *UserService {
	return &UserService{userRepo: userRepo, roleRepo: roleRepo}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.userRepo.FindByUsername(ctx, username)
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *UserService) CreateUser(ctx context.Context, username, password, email, roleID string) (*User, error) {
	existingUser, _ := s.userRepo.FindByUsername(ctx, username)
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	if roleID != "" {
		_, err := s.roleRepo.FindByID(ctx, roleID)
		if err != nil {
			return nil, ErrRoleNotFound
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  string(hashedPassword),
		Email:     email,
		RoleID:    roleID,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
	existingUser, _ := s.userRepo.FindByID(ctx, user.ID)
	if existingUser == nil {
		return ErrUserNotFound
	}

	if user.Username != existingUser.Username {
		checkUser, _ := s.userRepo.FindByUsername(ctx, user.Username)
		if checkUser != nil && checkUser.ID != user.ID {
			return ErrUsernameExists
		}
	}

	if user.RoleID != "" {
		_, err := s.roleRepo.FindByID(ctx, user.RoleID)
		if err != nil {
			return ErrRoleNotFound
		}
	}

	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (*User, error) {
	return s.Authenticate(ctx, username, password)
}

func (s *UserService) Register(ctx context.Context, username, password, email, roleName string) (*User, error) {
	var roleID string
	if roleName != "" {
		role, err := s.roleRepo.FindByName(ctx, roleName)
		if err != nil {
			return nil, ErrRoleNotFound
		}
		roleID = role.ID
	}

	return s.CreateUser(ctx, username, password, email, roleID)
}

func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	return s.userRepo.Count(ctx)
}
