package mypod

import "testing"

func TestEpisodeThumbnailPath(t *testing.T) {
	testcases := []struct {
		thumbnails  []string
		episodePath string
		expected    string
	}{
		{
			thumbnails:  []string{},
			episodePath: "A Very Complicated Episode (Name).m4a",
			expected:    "",
		},
		{
			thumbnails:  []string{"A Very Complicated Episode (Name).webp"},
			episodePath: "A Very Complicated Episode (Name).m4a",
			expected:    "/images/A Very Complicated Episode (Name).webp",
		},
		{
			thumbnails:  []string{"A Very, Very Complicated Episode (Name).jpg", "A Very Complicated Episode (Name).webp"},
			episodePath: "A Very Complicated Episode (Name).m4a",
			expected:    "/images/A Very Complicated Episode (Name).webp",
		},
	}
	for _, testcase := range testcases {
		actual := episodeThumbnailPath(testcase.thumbnails, testcase.episodePath)
		if actual != testcase.expected {
			t.Fatalf("fetched incorrect thumbnail path: expected %q, got: %q", testcase.expected, actual)
		}
	}
}
