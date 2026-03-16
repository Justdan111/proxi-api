package auth

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/crypto/bcrypt"

    "github.com/Justdan111/proxi-api/internal/user"
)

var (
    ErrEmailTaken    = errors.New("email already in use")
    ErrInvalidCreds  = errors.New("invalid email or password")
    ErrUserNotFound  = errors.New("user not found")
)

type Claims struct {
    UserID string `json:"userId"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

type Service struct {
    userRepo  *user.Repository
    jwtSecret string
    jwtExpiry int
}

func NewService(userRepo *user.Repository, jwtSecret string, jwtExpiry int) *Service {
    return &Service{
        userRepo:  userRepo,
        jwtSecret: jwtSecret,
        jwtExpiry: jwtExpiry,
    }
}

type SignupInput struct {
    Name     string `json:"name"     validate:"required,min=2"`
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type LoginInput struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthResult struct {
    Token string           `json:"token"`
    User  user.UserResponse `json:"user"`
}

func (s *Service) Signup(ctx context.Context, input SignupInput) (*AuthResult, error) {
    // Check if email already exists
    existing, err := s.userRepo.FindByEmail(ctx, input.Email)
    if existing != nil {
        return nil, ErrEmailTaken
    }
    if err != nil && err != mongo.ErrNoDocuments {
        return nil, err
    }

    // Hash password
    hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    newUser := &user.User{
        Name:     input.Name,
        Email:    input.Email,
        Password: string(hashed),
    }

    if err := s.userRepo.Create(ctx, newUser); err != nil {
        return nil, err
    }

    token, err := s.generateToken(newUser)
    if err != nil {
        return nil, err
    }

    return &AuthResult{Token: token, User: newUser.ToResponse()}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
    u, err := s.userRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, ErrInvalidCreds // Don't reveal whether email exists
    }

    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.Password)); err != nil {
        return nil, ErrInvalidCreds
    }

    token, err := s.generateToken(u)
    if err != nil {
        return nil, err
    }

    return &AuthResult{Token: token, User: u.ToResponse()}, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (*user.UserResponse, error) {
    // Parse ObjectID from string
    oid, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, ErrUserNotFound
    }

    u, err := s.userRepo.FindByID(ctx, oid)
    if err != nil {
        return nil, ErrUserNotFound
    }

    resp := u.ToResponse()
    return &resp, nil
}

func (s *Service) generateToken(u *user.User) (string, error) {
    claims := Claims{
        UserID: u.ID.Hex(),
        Email:  u.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpiry) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(s.jwtSecret), nil
    })

    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }

    claims, ok := token.Claims.(*Claims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }
    return claims, nil
}