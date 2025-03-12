package quora

import (
	"context"
	"os"
	"strings"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/client"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

func QuoraByheadless(websiteURL string) string {
	var allocCtx context.Context
	var cancelCtx context.CancelFunc
	allocOpts := chromedp.DefaultExecAllocatorOptions[:]
	allocOpts = append(allocOpts,
		chromedp.DisableGPU,
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("no-first-run", true),
		//chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36`),
		//chromedp.Flag("accept-language", `zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,zh-TW;q=0.6`),
	)

	headlessSer := os.Getenv("HEADLESS_SERVER_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if headlessSer != "" {
		c, cancelAlloc := chromedp.NewRemoteAllocator(ctx, headlessSer)
		defer cancelAlloc()
		allocCtx, cancelCtx = chromedp.NewContext(c)
	} else {
		c, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
		defer cancelAlloc()

		allocCtx, cancelCtx = chromedp.NewContext(c)
	}

	urlDomain, urlPrimaryDomain := client.GetPrimaryDomain(websiteURL)
	domainList := client.LoadCookieInfoManager(urlDomain, urlPrimaryDomain)
	var cookies []*network.CookieParam
	for _, domain := range domainList {
		for _, record := range domain.Records {
			if strings.HasPrefix(record.Domain, ".") {
				if len(record.Domain)-len(urlDomain) > 1 {
					continue
				}
			} else {
				if record.Domain != urlDomain {
					continue
				}
			}
			cookieVal := record.Value
			cookie := &network.CookieParam{
				Name:   record.Name,
				Value:  cookieVal,
				Path:   record.Path,
				Domain: record.Domain,
			}
			cookies = append(cookies, cookie)
		}
	}
	//ctx, cancel := chromedp.NewContext(context.Background())
	defer cancelCtx()
	htmlContent := ""
	common.Logger.Info("notion headless fetch 1 ")
	var lh, nh int64
	err := chromedp.Run(allocCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, cookie := range cookies {
				if err := network.SetCookie(cookie.Name, cookie.Value).WithDomain(cookie.Domain).WithPath(cookie.Path).Do(ctx); err != nil {
					return err
				}
			}
			return nil
		}),
		chromedp.Navigate(websiteURL),
		chromedp.Evaluate(`document.body.scrollHeight`, &lh),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for {
				if err := chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil).Do(ctx); err != nil {
					return err
				}
				time.Sleep(1 * time.Second)
				if err := chromedp.Evaluate(`document.body.scrollHeight`, &nh).Do(ctx); err != nil {
					return err
				}
				if nh == lh {
					break
				}
				lh = nh
			}
			return nil
		}),
		//chromedp.WaitVisible(`div.PageWrapper`, chromedp.ByQuery),
		/*chromedp.ActionFunc(func(ctx context.Context) error {
			for {
				if err := chromedp.Evaluate(`document.readyState`, &readyState).Do(ctx); err != nil {
					return err
				}
				if readyState == "complete" {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			return nil
		}),*/
		//chromedp.Poll(`document.readyState === 'complete'`, nil),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		common.Logger.Error("threads headless fetch error", zap.String("url", websiteURL), zap.Error(err))
	}
	common.Logger.Info("threads headless fetch end...", zap.Int("content len", len(htmlContent)))

	/*fileWriteErr := os.WriteFile("quota.txt", []byte(htmlContent), 0644)
	if fileWriteErr != nil {
		fmt.Println("Error writing file:", fileWriteErr)
	}*/

	return htmlContent
}
