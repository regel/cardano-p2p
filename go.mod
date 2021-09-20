module github.com/regel/cardano-p2p

go 1.15

replace github.com/regel/cardano-p2p/server => ./server

replace github.com/regel/cardano-p2p/pkg => ./pkg

replace github.com/regel/cardano-p2p/log => ./log

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/go-redis/redis/v8 v8.11.3
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	go.uber.org/zap v1.17.0
	gopkg.in/validator.v1 v1.0.0-20140827164146-4379dff89709
)
