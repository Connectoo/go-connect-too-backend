package auth

func (s *Service) createLifecycleToken() (plain, hash string, err error) {
	plain, err = generateRefreshToken()
	if err != nil {
		return "", "", err
	}
	hash = hashLifecycleToken(plain, s.lifecycleSecret)
	return plain, hash, nil
}

func hashLifecycleToken(plain string, secret []byte) string {
	return hashRefreshToken(plain, secret)
}
