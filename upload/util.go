package upload

func pctMax(artifactSize, maxSize uint64) float64 {
	return float64(100.0) * (float64(artifactSize) / float64(maxSize))
}
