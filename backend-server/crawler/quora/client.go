package quora

import (
	"context"
	"fmt"
	"os"
	"time"

	"bytetrade.io/web3os/backend-server/common"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	common.Logger.Info("notion headless fetch 1 ")
	err := chromedp.Run(allocCtx,
		chromedp.SetHeader("Cookie", "m-b=7WhaFYLMio-kAv38IL9v2A==; m-b_lax=7WhaFYLMio-kAv38IL9v2A==; m-b_strict=7WhaFYLMio-kAv38IL9v2A==; m-s=gNSarWSVCk3HAaIqkNLjjw==; m-theme=light; m-dynamicFontSize=regular; m-themeStrategy=auto; m-login=1; m-uid=2938460259; _fbp=fb.1.1739193699689.69139755452247159; _scid=mKY-7G-zOKoiLK5uTW7hmHq7-mhs9koQ; _scid_r=mKY-7G-zOKoiLK5uTW7hmHq7-mhs9koQ; _gcl_au=1.1.276664614.1739193700; _sctr=1%7C1739116800000; _sc_cspv=https%3A%2F%2Ftr.snapchat.com%2Fp; __stripe_mid=25c35dd5-65c7-4b50-b406-96f8547c0c7f0c18c4; m-sa=1; __gads=ID=be033e02750af23c:T=1739193714:RT=1741136934:S=ALNI_MbQBq_39LnERLLQvaD-TJTd81U5Mw; __gpi=UID=000010409a447058:T=1739193714:RT=1741136934:S=ALNI_MbhlCLrLHZeOUFIEAcp-3SP2YzlFg; __eoi=ID=bc643b054122a0a1:T=1739193714:RT=1741136934:S=AA-AfjZI10MqDPZOh3Dy9YFEk5H5; m-screen_size=756x780"),
		chromedp.Navigate(websiteURL),
		chromedp.WaitVisible(`div.PageWrapper`, chromedp.ByQuery),
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

	fileWriteErr := os.WriteFile("quota.txt", []byte(htmlContent), 0644)
	if fileWriteErr != nil {
		fmt.Println("Error writing file:", fileWriteErr)
	}

	return htmlContent
}
