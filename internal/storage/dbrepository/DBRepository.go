package dbrepository

type MetricRepository struct{}

func (r *MetricRepository) Store() error {
	/*_, err := r.db.Exec("INSERT INTO urls (short_code, long_url) VALUES (?, ?)", shortCode, longURL)*/
	return nil
}

func (r *MetricRepository) Get() (string, error) {
	/*_, err := r.db.Exec("INSERT INTO urls (short_code, long_url) VALUES (?, ?)", shortCode, longURL)*/
	return "", nil
}
