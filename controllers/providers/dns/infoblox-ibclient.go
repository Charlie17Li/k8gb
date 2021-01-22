package dns

import (
	"fmt"
	"strconv"

	ibclient "github.com/infobloxopen/infoblox-go-client"
)

func (p *InfobloxProvider) infobloxConnection() (*ibclient.ObjectManager, error) {
	hostConfig := ibclient.HostConfig{
		Host:     p.config.Infoblox.Host,
		Version:  p.config.Infoblox.Version,
		Port:     strconv.Itoa(p.config.Infoblox.Port),
		Username: p.config.Infoblox.Username,
		Password: p.config.Infoblox.Password,
	}
	transportConfig := ibclient.NewTransportConfig("false", 20, 10)
	requestBuilder := &ibclient.WapiRequestBuilder{}
	requestor := &ibclient.WapiHttpRequestor{}

	var objMgr *ibclient.ObjectManager

	if p.config.Override.FakeInfobloxEnabled {
		fqdn := "fakezone.example.com"
		fakeRefReturn := "zone_delegated/ZG5zLnpvbmUkLl9kZWZhdWx0LnphLmNvLmFic2EuY2Fhcy5vaG15Z2xiLmdzbGJpYmNsaWVudA:fakezone.example.com/default"
		ohmyFakeConnector := &fakeInfobloxConnector{
			getObjectObj: ibclient.NewZoneDelegated(ibclient.ZoneDelegated{Fqdn: fqdn}),
			getObjectRef: "",
			resultObject: []ibclient.ZoneDelegated{*ibclient.NewZoneDelegated(ibclient.ZoneDelegated{Fqdn: fqdn, Ref: fakeRefReturn})},
		}
		objMgr = ibclient.NewObjectManager(ohmyFakeConnector, "ohmyclient", "")
	} else {
		conn, err := ibclient.NewConnector(hostConfig, transportConfig, requestBuilder, requestor)
		if err != nil {
			return nil, err
		}
		defer func() {
			err = conn.Logout()
			if err != nil {
				p.assistant.Error(err, "Failed to close connection to infoblox")
			}
		}()
		objMgr = ibclient.NewObjectManager(conn, "ohmyclient", "")
	}
	return objMgr, nil
}

func (p *InfobloxProvider) checkZoneDelegated(findZone *ibclient.ZoneDelegated) error {
	if findZone.Fqdn != p.config.DNSZone {
		err := fmt.Errorf("delegated zone returned from infoblox(%s) does not match requested gslb zone(%s)", findZone.Fqdn, p.config.DNSZone)
		return err
	}
	return nil
}

func (p *InfobloxProvider) filterOutDelegateTo(delegateTo []ibclient.NameServer, fqdn string) []ibclient.NameServer {
	for i := 0; i < len(delegateTo); i++ {
		if delegateTo[i].Name == fqdn {
			delegateTo = append(delegateTo[:i], delegateTo[i+1:]...)
			i--
		}
	}
	return delegateTo
}
