package mitm

import "fmt"

func Example() {
	certPEM, keyPEM, err := GenerateDevCA("proxykit dev ca", 1)
	if err != nil {
		panic(err)
	}
	authority, err := LoadAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	policy := Policy{
		Authority:   authority,
		AllowSuffix: []string{".example.com"},
		DenySuffix:  []string{"blocked.example.com"},
	}

	leaf, err := authority.IssueFor("api.example.com:443")
	if err != nil {
		panic(err)
	}

	fmt.Println(policy.ShouldIntercept("api.example.com:443"))
	fmt.Println(policy.ShouldIntercept("blocked.example.com"))
	fmt.Println(len(leaf.Certificate) > 0)
	// Output:
	// true
	// false
	// true
}
