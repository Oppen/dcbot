# DCUBABOT

Este proyecto es una reescritura en Go de [dcubabot](https://github.com/rozen03/dcubabot),
autoría mayormente de Rozen.

La principal razón es que estaba aburrido, me gusta Go, quería escribir un bot
y no se me caía una idea respecto a qué (firma: Mario).

Por otro lado, se planea extender un poco, además de optimizar el rendimiento.
Las features me parecen bien justificadas (ver issues), lo del rendimiento es
nuevamente porque pintó. Ni siquiera voy a medirlo :)

Para ejecutarlo localmente basta con instalar Go y ejecutar `go run`.

El nuevo bot está preparado para correr fácilmente en Docker, usando el
Dockerfile y creando ciertos archivos de configuración para el runtime.
El mismo se encarga de compilar el programa.

[//]: # (TODO: expandir esto!)
