### docker pull latala/crawler-cli:latest
### https://hub.docker.com/r/latala/crawler-cli/tags

# Тестовое задание для Golang разработчика

Написать консольную утилиту - crawler, которая будет рекурсивно парсить страницы, находить в них ссылки на другие (ранее не спаршенные) страницы и тоже их парсить.

## Параметры для запуска

* `n` - максимальное количество параллельных запросов

* `root` - начальная страница

* `r` - глубина рекурсии

* `user-agent` - заголовок `User-Agent`

Например: `./crawler -r 10 -root https://en.wikipedia.org/wiki/Money_Heist -n 15`

## Ожидаемый результат

Вывести в stdout количество уникальных страниц.

Затем вывести построчно все посещенные страницы с указанием глубины рекурсии.

Пример:

```txt
4
1 https://mysite.com
2 https://mysite.com/help
2 https://mysite.com/about
3 https://mysite.com/about/contacts
```

## Решение

Решение оформить в виде docker image на базе alpine. Запушить его в docker.hub и указать его в readme.md