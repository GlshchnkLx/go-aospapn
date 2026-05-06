package main

import (
	"fmt"
	"io"
)

func usage(writer io.Writer) {
	fmt.Fprintln(writer, `Usage:
  apnctl fetch    --out apns-full-conf.xml
  apnctl stats    --in apns-full-conf.xml
  apnctl list     --in apns-full-conf.xml --kind plmn
  apnctl find     --in apns-full-conf.xml --plmn 25001 --type default --output-format table
  apnctl convert  --in apns-full-conf.xml --output-format json
  apnctl patch    --in apns-full-conf.xml --plmn 25001 --set base.profileID=42
  apnctl validate --in apns-full-conf.xml --strict
  apnctl inspect  --in apns-full-conf.xml --plmn 25001
  apnctl build    --carrier "Operator" --mcc 999 --mnc 99 --apn iot.example

Input flags: --in, --stdin, --url, --base64, --input-format xml|json
Output flags: --out, --output-format xml|json|table|csv|text|summary, --flat, --group-by, --dedupe-by, --normalize, --offset, --limit
Filter flags: --plmn, --mcc, --mnc, --carrier-id, --carrier, --apn, --apn-contains, --type, --protocol, --network, --has, --without, --valid-only, --invalid-only, --not`)
}
