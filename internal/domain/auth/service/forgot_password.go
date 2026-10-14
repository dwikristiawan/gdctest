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

func (s *AuthService) ForgotPassword(ctx context.Context, email, ip string) error {
	// 1. Rate Limiter check (Max 3 attempts, Ban 30 mins)
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

	// 2. Cek apakah user terdaftar
	existingUser, err := s.repo.FindOne(ctx, model.User{Email: email})
	if err != nil {
		return apperror.New(http.StatusInternalServerError, "database error: "+err.Error())
	}
	if existingUser == nil {
		return apperror.New(http.StatusBadRequest, "email not found")
	}

	// 3. Generate token
	tok, err := token.GenerateRegistrationToken(email)
	if err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to generate token: "+err.Error())
	}

	// 4. Simpan ke Redis dengan TTL 15 MENIT (overwrite link lama)
	session := &model.RegistrationSession{
		Token: tok,
		IP:    ip,
	}
	if err := s.repo.SaveResetPasswordSession(ctx, email, session, 15*time.Minute); err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to save reset password session: "+err.Error())
	}

	// 5. Kirim Email Reset Link
	if s.mailer != nil {
		link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.BaseURL, tok)
		htmlBody := fmt.Sprintf(`
			<h2>Reset Password Request</h2>
			<p>Click the link below to reset your password (valid for 15 minutes):</p>
			<p><a href="%s">%s</a></p>
			<p>Notice: Requesting a new link will invalidate this one.</p>
		`, link, link)

		_ = s.mailer.SendEmail(email, "Reset Your Password", htmlBody)
	}

	return nil
}
