package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/starudream/go-lib/cobra/v2"
	"github.com/starudream/go-lib/core/v2/slog"
	"github.com/starudream/go-lib/cron/v2"
	"github.com/starudream/go-lib/ntfy/v2"
	"github.com/starudream/go-lib/service/v2"

	"github.com/starudream/miyoushe-task/config"
	"github.com/starudream/miyoushe-task/job"
)

var cronCmd = cobra.NewCommand(func(c *cobra.Command) {
	c.Use = "cron"
	c.Short = "Run as cron job"
	c.RunE = func(cmd *cobra.Command, args []string) error {
		return service.New("miyoushe-task", nil).Run()
	}
})

func init() {
	rootCmd.AddCommand(cronCmd)
	rand.Seed(time.Now().UnixNano()) // 初始化随机种子
}

func randomDuration(maxOffsetHours int) time.Duration {
	offset := rand.Intn(maxOffsetHours * 3600) // 生成随机秒数
	return time.Duration(offset) * time.Second
}

func cronRun() error {
	if config.C().Cron.Startup {
		cronJob()
	}
	err := cron.AddJob(config.C().Cron.Spec, "miyoushe-cron", func() {
		delay := randomDuration(6) // 生成一个最大为6小时的随机延迟
		time.Sleep(delay)          // 等待计算出的延迟时间
		cronJob()
	})
	if err != nil {
		return fmt.Errorf("add cron job error: %w", err)
	}
	cron.Run()
	return nil
}

func cronJob() {
	for i := 0; i < len(config.C().Accounts); i++ {
		cronForumAccount(config.C().Accounts[i])
		cronGameAccount(config.C().Accounts[i])
	}
}

func cronForumAccount(account config.Account) (msg string) {
	record, err := job.SignForum(account)
	if err != nil {
		msg = fmt.Sprintf("%s: %v", record.Name(), err)
		slog.Error(msg)
	} else {
		msg = account.Phone + " " + record.Success()
		slog.Info(msg)
	}
	err = ntfy.Notify(context.Background(), msg)
	if err != nil && !errors.Is(err, ntfy.ErrNoConfig) {
		slog.Error("cron notify error: %v", err)
	}
	return
}

func cronGameAccount(account config.Account) (msg string) {
	records, err := job.SignGame(account)
	if err != nil {
		msg = fmt.Sprintf("%s: %v", records.Name(), err)
		slog.Error(msg)
	} else {
		msg = account.Phone + " " + records.Success()
		slog.Info(msg)
	}
	err = ntfy.Notify(context.Background(), msg)
	if err != nil && !errors.Is(err, ntfy.ErrNoConfig) {
		slog.Error("cron notify error: %v", err)
	}
	return
}
