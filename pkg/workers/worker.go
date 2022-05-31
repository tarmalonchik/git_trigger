package serversPing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Alan-prog/wireguard_vpn/internal/app/master/config"
	"github.com/Alan-prog/wireguard_vpn/internal/app/master/model/masterModel"
	"github.com/Alan-prog/wireguard_vpn/internal/pkg/hezner"
	"github.com/sirupsen/logrus"
)

const (
	slaveTempName    = "slave-%d"
	upStatus         = "up"
	masterServerName = "master-server"
)

type Worker struct {
	conf         *config.Config
	model        *masterModel.Model
	httpClient   *http.Client
	heznerClient *hezner.Client
}

func NewServersPingWorker(conf *config.Config, model *masterModel.Model, heznerClient *hezner.Client) *Worker {
	svcControllerWorker := &Worker{
		conf:  conf,
		model: model,
	}
	svcControllerWorker.heznerClient = heznerClient
	svcControllerWorker.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	return svcControllerWorker
}

func (t *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logrus.WithError(ctx.Err()).Info("stop serversPing worker")
			return nil
		case <-time.NewTicker(5 * time.Second).C:
			if err := t.Process(ctx); err != nil {
				logrus.Errorf("serversPing.Run error while process: %v", err)
			}
		}
	}
}

func (t *Worker) Process(ctx context.Context) error {
	servers, err := t.heznerClient.Cloud.GetAllServers(ctx)
	if err != nil {
		return fmt.Errorf("serversPing.Process error while getting servers from hezner: %w", err)
	}

	serversIPsNeedToCheck := make([]*hezner.ServerItem, 0)

	for i := range servers.Servers {
		if servers.Servers[i].Name == masterServerName {
			continue
		}
		if len(servers.Servers[i].PrivateNet) != 0 {
			if len(servers.Servers[i].PrivateNet) > 1 {
				logrus.Errorf("bad count of private network")
				continue
			}
			serversIPsNeedToCheck = append(serversIPsNeedToCheck, &servers.Servers[i])
		}
	}

	for i := range serversIPsNeedToCheck {
		resp, err := t.sendHealthRequest(ctx, serversIPsNeedToCheck[i].PrivateNet[0].Ip)
		if err != nil {
			return fmt.Errorf("serversPing.Process error while sending health req: %w", err)
		}
		if resp != upStatus {
			logrus.Errorf("serversPing.Process error status from server with ip: %s",
				serversIPsNeedToCheck[i].PrivateNet[0].Ip)
			continue
		}
	}

	if err := t.checkZonesAndUpdateIfNeed(ctx, serversIPsNeedToCheck); err != nil {
		return fmt.Errorf("serversPing.Process error checking dns zones: %w", err)
	}

	if err := t.checkUsersCountAndAddServersIfNeed(ctx, int64(len(servers.Servers))); err != nil {
		return fmt.Errorf("serversPing.Process error adding new servers: %w", err)
	}
	return nil
}

func (t *Worker) checkUsersCountAndAddServersIfNeed(ctx context.Context, allServersCount int64) error {
	usersCount, err := t.model.GetActiveUsersCount(ctx)
	if err != nil {
		return fmt.Errorf("serversPing.checkUsersCountAndAddServersIfNeed error getting users count: %w", err)
	}

	if (allServersCount * t.conf.MaxCountPerServer) < usersCount {
		if _, err := t.heznerClient.Cloud.CreateNewServer(ctx, fmt.Sprintf(slaveTempName, time.Now().Unix())); err != nil {
			return fmt.Errorf("serversPing.checkUsersCountAndAddServersIfNeed error creating server: %w", err)
		}
	}
	return nil
}

func (t *Worker) checkZonesAndUpdateIfNeed(ctx context.Context, ipsNeedToAddDns []*hezner.ServerItem) error {
	zones, err := t.heznerClient.Dns.GetAllZones(ctx)
	if err != nil {
		return fmt.Errorf("serversPing.checkZonesAndUpdateIfNeed error while getting zones: %w", err)
	}

	zonesMap := make(map[string]interface{})

	for i := range zones.Records {
		zonesMap[zones.Records[i].Value] = nil
	}

	for i := range ipsNeedToAddDns {
		if _, ok := zonesMap[ipsNeedToAddDns[i].PublicNet.Ipv4.Ip]; !ok {
			if err := t.heznerClient.Dns.CreateARecordRequest(ctx, ipsNeedToAddDns[i].PublicNet.Ipv4.Ip); err != nil {
				logrus.Errorf("serversPing.checkZonesAndUpdateIfNeed error while adding zone: %v", err)
				continue
			}
		}
	}
	return nil
}

func (t *Worker) sendHealthRequest(ctx context.Context, ip string) (string, error) {
	var resp healthResponse

	req, err := http.NewRequest(
		http.MethodGet,
		"http://"+ip+":8080/_healthz",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("serversPing.Process error while creating http req: %w", err)
	}

	httpResp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("serversPing.Process error during requesting: %w", err)
	}

	bodyByte, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("serversPing.Process error during reading body: %w", err)
	}

	if err := json.Unmarshal(bodyByte, &resp); err != nil {
		return "", fmt.Errorf("serversPing.Process error during unmarshal response: %w", err)
	}

	return resp.Status, nil
}
