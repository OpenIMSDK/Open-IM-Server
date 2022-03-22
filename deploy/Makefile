
GREEN_PREFIX="\033[32m"
COLOR_SUFFIX="\033[0m"
SKY_BLUE_PREFIX="\033[36m"


# 编译所有需要的组件源码
win-build-all:
	go env -w GOOS=linux 
	make build-api && make build-msg-gateway && make build-msg-transfer  && make build-push && make build-timer-task
	make build-rpc-user && make build-rpc-friend && make build-rpc-group && make build-rpc-msg && make build-rpc-auth
	make build-demo
	go env -w GOOS=windows

# 编译 open_im_api
build-api:
	echo -e ${GREEN_PREFIX} "open_im_api building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_api ../cmd/open_im_api/main.go
	echo -e ${GREEN_PREFIX} "open_im_api build ok" ${COLOR_SUFFIX}

# 编译 open_im_msg_gateway
build-msg-gateway:
	echo -e ${GREEN_PREFIX} "open_im_msg_gateway building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_msg_gateway ../cmd/open_im_msg_gateway/main.go
	echo -e ${GREEN_PREFIX} "open_im_msg_gateway build ok" ${COLOR_SUFFIX}

# 编译 open_im_msg_transfer
build-msg-transfer:
	echo -e ${GREEN_PREFIX} "open_im_msg_transfer building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_msg_transfer ../cmd/open_im_msg_transfer/main.go
	echo -e ${GREEN_PREFIX} "open_im_msg_transfer build ok" ${COLOR_SUFFIX}

# 编译 open_im_push
build-push:
	echo -e ${GREEN_PREFIX} "open_im_push building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_push ../cmd/open_im_push/main.go
	echo -e ${GREEN_PREFIX} "open_im_push build ok" ${COLOR_SUFFIX}

# 编译 open_im_timer_task
build-timer-task:
	echo -e ${GREEN_PREFIX} "open_im_timer_task building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_timer_task ../cmd/open_im_timer_task/main.go
	echo -e ${GREEN_PREFIX} "open_im_timer_task build ok" ${COLOR_SUFFIX}

# 编译 build-rpc-user
build-rpc-user:
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_user building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_rpc_user ../cmd/rpc/open_im_user/main.go
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_user build ok" ${COLOR_SUFFIX}

# 编译 build-rpc-friend
build-rpc-friend:
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_friend building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_rpc_friend ../cmd/rpc/open_im_friend/main.go
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_friend build ok" ${COLOR_SUFFIX}

# 编译 build-rpc-group
build-rpc-group:
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_group building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_rpc_group ../cmd/rpc/open_im_group/main.go
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_group build ok" ${COLOR_SUFFIX}

# 编译 build-rpc-auth
build-rpc-auth:
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_auth building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_rpc_auth ../cmd/rpc/open_im_auth/main.go
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_auth build ok" ${COLOR_SUFFIX}

# 编译 build-rpc-msg
build-rpc-msg:
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_msg building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_rpc_msg ../cmd/rpc/open_im_msg/main.go
	echo -e ${SKY_BLUE_PREFIX} "open_im_rpc_msg build ok" ${COLOR_SUFFIX}

# 编译 open_im_demo
build-demo:
	echo -e ${SKY_BLUE_PREFIX} "open_im_demo building..." ${COLOR_SUFFIX}
	go build -ldflags="-w -s" -o open_im_demo ../cmd/open_im_demo/main.go
	echo -e ${SKY_BLUE_PREFIX} "open_im_demo build ok" ${COLOR_SUFFIX}

# 打包所有组件为镜像
image-all:
	make image-api && make image-msg-gateway && make image-msg-transfer & make image-push && make image-timer-task
	make image-rpc-user && make image-rpc-friend && make image-rpc-group && make image-rpc-msg && make image-rpc-auth
	make image-demo

# 打包 open_im_api
image-api:
	echo -e ${GREEN_PREFIX} "IMAGE:openim/api building..." ${COLOR_SUFFIX}
	docker build -t openim/api:latest -f ./dockerfiles/Dockerfile.api .
	echo -e ${GREEN_PREFIX} "IMAGE:openim/api build ok" ${COLOR_SUFFIX}

# 打包 open_im_msg_gateway
image-msg-gateway:
	echo -e ${GREEN_PREFIX} "IMAGE:openim/msg_gateway building..." ${COLOR_SUFFIX}
	docker build -t openim/msg_gateway:latest -f ./dockerfiles/Dockerfile.msg_gateway .
	echo -e ${GREEN_PREFIX} "IMAGE:openim/msg_gateway build ok" ${COLOR_SUFFIX}

# 打包 open_im_msg_transfer
image-msg-transfer:
	echo -e ${GREEN_PREFIX} "IMAGE:openim/msg_transfer building..." ${COLOR_SUFFIX}
	docker build -t openim/msg_transfer:latest -f ./dockerfiles/Dockerfile.msg_transfer .
	echo -e ${GREEN_PREFIX} "IMAGE:openim/msg_transfer build ok" ${COLOR_SUFFIX}

# 打包 open_im_push
image-push:
	echo -e ${GREEN_PREFIX} "IMAGE:openim/push building..." ${COLOR_SUFFIX}
	docker build -t openim/push:latest -f ./dockerfiles/Dockerfile.push .
	echo -e ${GREEN_PREFIX} "IMAGE:openim/push build ok" ${COLOR_SUFFIX}

# 打包 open_im_timer_task
image-timer-task:
	echo -e ${GREEN_PREFIX} "IMAGE:openim/timer_task building..." ${COLOR_SUFFIX}
	docker build -t openim/timer_task:latest -f ./dockerfiles/Dockerfile.timer_task .
	echo -e ${GREEN_PREFIX} "IMAGE:openim/timer_task build ok" ${COLOR_SUFFIX}

# 打包 build-rpc-user
image-rpc-user:
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_user building..." ${COLOR_SUFFIX}
	docker build -t openim/rpc_user:latest -f ./dockerfiles/Dockerfile.rpc_user .
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_user build ok" ${COLOR_SUFFIX}

# 打包 build-rpc-friend
image-rpc-friend:
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_friend building..." ${COLOR_SUFFIX}
	docker build -t openim/rpc_friend:latest -f ./dockerfiles/Dockerfile.rpc_friend .
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_friend build ok" ${COLOR_SUFFIX}

# 打包 build-rpc-group
image-rpc-group:
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_group building..." ${COLOR_SUFFIX}
	docker build -t openim/rpc_group:latest -f ./dockerfiles/Dockerfile.rpc_group .
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_group build ok" ${COLOR_SUFFIX}

# 打包 build-rpc-auth
image-rpc-auth:
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_auth building..." ${COLOR_SUFFIX}
	docker build -t openim/rpc_auth:latest -f ./dockerfiles/Dockerfile.rpc_auth .
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_auth build ok" ${COLOR_SUFFIX}

# 打包 build-rpc-msg
image-rpc-msg:
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_msg building..." ${COLOR_SUFFIX}
	docker build -t openim/rpc_msg:latest -f ./dockerfiles/Dockerfile.rpc_msg .
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/rpc_msg build ok" ${COLOR_SUFFIX}

# 打包 open_im_demo
image-demo:
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/demo building..." ${COLOR_SUFFIX}
	docker build -t openim/demo:latest -f ./dockerfiles/Dockerfile.demo .
	echo -e ${SKY_BLUE_PREFIX} "IMAGE:openim/demo build ok" ${COLOR_SUFFIX}

.PHONY: win-build-all build-api build-msg-gateway build-msg-transfer build-push 
	build-timer-task build-rpc-user build-rpc-friend build-rpc-group build-rpc-msg build-demo
	image-all image-api image-msg-gateway image-msg-transfer image-push 
	image-timer-task image-rpc-user image-rpc-friend image-rpc-group image-rpc-msg image-demo
