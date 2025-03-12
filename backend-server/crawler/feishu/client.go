package feishu

import (
	"context"
	"fmt"
	"os"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

func FeishuByheadless(websiteURL string) string {
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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
	//ctx, cancel := chromedp.NewContext(context.Background())
	defer cancelCtx()
	htmlContent := ""
	common.Logger.Info("feishu headless fetch 1 ")
	//var isHidden bool
	err := chromedp.Run(allocCtx,
		chromedp.Navigate(websiteURL),
		//chromedp.WaitVisible(`div.page-block`, chromedp.ByQuery),
		chromedp.WaitNotPresent("span.grid-column-percent", chromedp.ByQuery),

		/*chromedp.Evaluate(`
			const element = document.querySelector('span.grid-column-percent');
			if (!element) {
				true;
			} else {
				const style = window.getComputedStyle(element);
				style.display === 'none' || style.visibility === 'hidden' || element.offsetWidth === 0 || element.offsetHeight === 0;
			}
		`, &isHidden),*/
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		common.Logger.Error("feishu headless fetch error", zap.String("url", websiteURL), zap.Error(err))
	}
	common.Logger.Info("feishu headless fetch end...", zap.Int("content len", len(htmlContent)))

	fileWriteErr := os.WriteFile("feishu.html", []byte(htmlContent), 0644)
	if fileWriteErr != nil {
		fmt.Println("Error writing file:", fileWriteErr)
	}

	return htmlContent
}
