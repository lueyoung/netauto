include Makefile.inc

define log 
	@echo "`date` - [INFO] - $(1)"
endef

all: rm cp compile mv

rm:
	@rm -rf ${GOPATH}/*

cp:
	@mkdir -p ${GOPATH} 
	@yes | cp ${SRC}/* ${GOPATH} 

compile:
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./main.go

mv:
	@mkdir -p ${BIN} 
	@yes | mv main ${BIN}
	@yes | cp ${BIN}/main /tmp
	@scp ${BIN}/main root@192.168.100.161:/tmp

deploy: all run

run:
	@ansible ${VM} -m copy -a "src=${BIN}/main dest=${DEST_PATH} mode='a+x'" 
	@ansible ${VM} -m copy -a "src=${MANIFEST}/core.service dest=/etc/systemd/system/" 
	@ansible ${VM} -m shell -a "systemctl daemon-reload"
	@ansible ${VM} -m shell -a "systemctl enable core"
	@ansible ${VM} -m shell -a "systemctl restart core"

sync:
	@git pull origin `git branch | grep "*" | tr -d "* "`

test-sync:
	@echo "git pull origin `git branch | grep '*' | tr -d '* '`"

docker: docker-run docker-cp docker-compile docker-export docker-clear

docker-run:
	$(call log, "start a docker container")
	@docker ps | grep ${NAME} > /dev/nul 2>&1 || docker run -d --name ${NAME} -v `pwd`/src:/src ${IMAGE} tail -f /dev/null 
	@docker ps | grep ${NAME} > /dev/nul 2>&1 && echo "`date` - [WARNING] - container ${NAME} already exists"  

docker-cp:
	$(call log, "copy source code to the container")
	@docker exec -it ${NAME} mkdir -p /workspace 
	@docker cp main.go ${NAME}:/workspace
	@chmod +x scripts/compile.sh
	@docker cp scripts/compile.sh ${NAME}:/workspace

docker-compile:
	$(call log, "compile the code")
	@docker exec -it ${NAME} /workspace/compile.sh 

docker-export:
	$(call log, "copy compiled file to directory bin")
	@mkdir -p bin
	@docker cp ${NAME}:/workspace/main bin
	@chmod +x bin/main 

docker-clear:
	$(call log, "remove the container")
	@docker stop ${NAME}
	@docker rm -f  ${NAME}

enter:
	@docker exec -it ${NAME} /bin/bash

test:
	$(call log, test)
