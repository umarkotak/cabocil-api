run:
	go run .

migrate-up:
	go run . migrate up

bin:
	go build -o cabocil-api .

bin_run:
	./cabocil-api

nohup_run:
	nohup ./cabocil-api &

stopd:
	pkill cabocil-api

statusd:
	ps aux | grep cabocil-api

logs:
	tail -f cabocil-api.error.log

install-service:
	sudo chmod +x cabocil-api
	sudo cp com.cabocil-api.plist /Library/LaunchDaemons
	sudo chmod +x /Library/LaunchDaemons/com.cabocil-api.plist
	sudo launchctl bootstrap system /Library/LaunchDaemons/com.cabocil-api.plist

uninstall-service:
	sudo launchctl unload /Library/LaunchDaemons/com.cabocil-api.plist
	sudo rm /Library/LaunchDaemons/com.cabocil-api.plist

start:
	sudo launchctl start com.cabocil-api

stop:
	sudo launchctl stop com.cabocil-api

deploy:
	git pull --rebase origin master
	make bin
	make stop
	make start

status:
	sudo lsof -i :33000
