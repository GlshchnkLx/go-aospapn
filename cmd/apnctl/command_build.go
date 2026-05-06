package main

import (
	"fmt"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func runBuild(args []string) error {
	flags, fs := newCommonFlagSet("build")
	var carrier, apn, apnType, protocol, roamingProtocol, network string
	var mcc, mnc, carrierID, profileID, mtu int
	var enabled, visible, editable string
	fs.StringVar(&carrier, "carrier", "", "carrier name")
	fs.IntVar(&carrierID, "carrier-id", -1, "carrier ID")
	fs.IntVar(&mcc, "mcc", -1, "MCC")
	fs.IntVar(&mnc, "mnc", -1, "MNC")
	fs.StringVar(&apn, "apn", "", "APN name")
	fs.StringVar(&apnType, "type", "default", "APN type")
	fs.IntVar(&profileID, "profile-id", -1, "profile ID")
	fs.StringVar(&protocol, "protocol", "", "bearer protocol")
	fs.StringVar(&roamingProtocol, "roaming-protocol", "", "roaming bearer protocol")
	fs.StringVar(&network, "network", "", "network bitmask")
	fs.IntVar(&mtu, "mtu", -1, "MTU")
	fs.StringVar(&enabled, "enabled", "", "carrier enabled: true or false")
	fs.StringVar(&visible, "visible", "", "user visible: true or false")
	fs.StringVar(&editable, "editable", "", "user editable: true or false")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if mcc < 0 || mnc < 0 || apn == "" {
		return fmt.Errorf("build requires --mcc, --mnc and --apn")
	}

	record := apnxml.Object{ObjectRoot: &apnxml.ObjectRoot{Carrier: carrier, Mcc: &mcc, Mnc: &mnc}}
	if carrierID >= 0 {
		record.CarrierID = &carrierID
	}
	base := apntool.EnsureBase(&record)
	base.Apn = &apn
	parsedType, err := apnxml.ParseObjectBaseType(apnType)
	if err != nil {
		return err
	}
	base.Type = &parsedType
	if profileID >= 0 {
		base.ProfileID = &profileID
	}
	if protocol != "" || roamingProtocol != "" || mtu >= 0 {
		bearer := apntool.EnsureBearer(&record)
		if protocol != "" {
			parsedProtocol, err := apnxml.ParseObjectBearerProtocol(protocol)
			if err != nil {
				return err
			}
			bearer.Type = &parsedProtocol
			if roamingProtocol == "" {
				bearer.TypeRoaming = &parsedProtocol
			}
		}
		if roamingProtocol != "" {
			parsedProtocol, err := apnxml.ParseObjectBearerProtocol(roamingProtocol)
			if err != nil {
				return err
			}
			bearer.TypeRoaming = &parsedProtocol
		}
		if mtu >= 0 {
			bearer.Mtu = &mtu
		}
	}
	if network != "" || enabled != "" || visible != "" || editable != "" {
		other := apntool.EnsureOther(&record)
		if network != "" {
			parsedNetwork, err := apnxml.ParseObjectNetworkType(network)
			if err != nil {
				return err
			}
			other.NetworkTypeBitmask = &parsedNetwork
		}
		if err := setOptionalBool(&other.CarrierEnabled, enabled); err != nil {
			return fmt.Errorf("--enabled: %w", err)
		}
		if err := setOptionalBool(&other.UserVisible, visible); err != nil {
			return fmt.Errorf("--visible: %w", err)
		}
		if err := setOptionalBool(&other.UserEditable, editable); err != nil {
			return fmt.Errorf("--editable: %w", err)
		}
	}
	record.Normalize()
	return writeAPNs(flags, apntool.From(apnxml.Array{record}))
}
