# dsh

Hassle-free Docker container shells.

## Setup

[Download the zero-install binary](https://github.com/schmich/dsh/releases) to a directory on your `PATH`.

## Usage

`dsh` simplifies `docker exec -it <container> sh`. It automatically determines the available
shells in a container and chooses the best one (e.g. `bash` > `sh`).

Run `dsh` by itself to get a list of running containers:

```
$ dsh
1  myapp_nginx-tls_1      1d70e94909a9  myapp_nginx-tls
2  myapp_varnish_1        421ac83e776e  myapp_varnish
3  myapp_nginx-app_1      54ce96eb48f3  myapp_nginx-app
4  myapp_app_1            5a305dbbc5be  myapp_app
5  myapp_laravel-queue_1  7afe8ea4c418  myapp_laravel-queue
6  myapp_solr_1           87077551c652  myapp_solr
7  myapp_mysql_1          6859adde40a3  myapp_mysql
8  myapp_redis_1          3c489fb03463  redis:3-alpine
9  myapp_beanstalkd_1     9c7cc9d4ae38  schickling/beanstalkd
a  myapp_artisan_run_1    cf45724062d0  myapp_artisan
b  myapp_gulp_run_2       1d2079ad1cf5  myapp_gulp
> 2

Running /bin/bash in myapp_varnish_1 (421ac83e776e).
[root@421ac83e776e /]# _
```

Run `dsh <query>` to search for containers:

```
$ dsh nginx
Multiple containers found for 'nginx'.
1  myapp_nginx-tls_1  1d70e94909a9  myapp_nginx-tls
2  myapp_nginx-app_1  54ce96eb48f3  myapp_nginx-app
> 1

Running /bin/ash in myapp_nginx-tls_1 (1d70e94909a9).
/etc/nginx # _
```

You can also search by container ID. If only one container is found, a shell is automatically started:

```
$ dsh 870
Running /bin/bash in myapp_solr_1 (87077551c652).
bash-4.3$
```

## License

Copyright &copy; 2016 Chris Schmich  
MIT License. See [LICENSE](LICENSE) for details.
