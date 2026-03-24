run:
	go run main.go --conns cmd://localhost/home/mawen/logs/app.log --log-level debug

run-mutli:
	go run main.go --conns cmd://localhost/home/mawen/logs/app.log,ssh://root:admin@192.168.122.6:22/home/root/logs/app.log
