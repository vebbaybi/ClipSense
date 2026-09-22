package fixture

import "crypto/tls"

func unsafe() *tls.Config {
 // ruleid: go-disabled-tls-validation
 return &tls.Config{InsecureSkipVerify: true}
}
