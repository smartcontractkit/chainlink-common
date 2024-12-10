package pg

import "net/url"

func SetApplicationName(u *url.URL, name string) {
	// trim to postgres limit
	name = name[:min(63, len(name))]
	q := u.Query()
	q.Set("application_name", name)
	u.RawQuery = q.Encode()
}
