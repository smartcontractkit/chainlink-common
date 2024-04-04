package pg

import "net/url"

func SetApplicationName(u *url.URL, name string) {
	// trim to postres limit
	if len(name) > 63 {
		name = name[:63]
	}
	q := u.Query()
	q.Set("application_name", name)
	u.RawQuery = q.Encode()
}
