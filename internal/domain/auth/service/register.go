package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/apperror"
	"gdctest/pkg/ratelimit"
	"gdctest/pkg/token"
)

func (s *AuthService) Register(ctx context.Context, email, ip string) error {
	// 1. Rate Limiter: Max 3 requests, Ban 30 mins
	if s.limiter != nil {
		res, err := s.limiter.Check(ctx, email, ratelimit.Config{
			MaxAttempts: 3,
			Window:      30 * time.Minute,
			BanDuration: 30 * time.Minute,
		})
		if err != nil {
			return apperror.New(http.StatusInternalServerError, "failed to check rate limit: "+err.Error())
		}
		if !res.Allowed {
			return apperror.New(http.StatusTooManyRequests, fmt.Sprintf("too many requests for this email. banned for %v", res.BanTTL.Round(time.Second)))
		}
	}

	// 2. Cek apakah email sudah terdaftar di PostgreSQL
	existingUser, err := s.repo.FindOne(ctx, model.User{Email: email})
	if err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to query user: "+err.Error())
	}
	if existingUser != nil {
		return apperror.New(http.StatusBadRequest, "email is already registered")
	}

	// 3. Generate registration token (payload encodes email)
	tok, err := token.GenerateRegistrationToken(email)
	if err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to generate token: "+err.Error())
	}

	// 4. Simpan ke Redis (overwrite existing session -> otomatis invalidasi link lama)
	session := &model.RegistrationSession{
		Token: tok,
		IP:    ip,
	}
	if err := s.repo.SaveRegistrationSession(ctx, email, session, 30*time.Minute); err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to save registration session: "+err.Error())
	}

	// 5. Kirim Verification Email
	if s.mailer != nil {
		link := fmt.Sprintf("%s/verify?token=%s", s.cfg.BaseURL, tok)
		htmlBody := fmt.Sprintf(`
			<h2>Verification Link</h2>
			<p>Click the link below to set your username and password (valid for 30 minutes):</p>
			<p><a href="%s">%s</a></p>
			<p>Notice: Requesting a new link will invalidate this one.</p>
		`, link, link)

		_ = s.mailer.SendEmail(email, "Complete Your Registration", htmlBody)
	}

	return nil
}
