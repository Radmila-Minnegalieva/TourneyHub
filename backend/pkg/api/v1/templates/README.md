# Шаблоны генератора

Шаблоны взяты из [oapi-codegen v2.6.0](https://github.com/oapi-codegen/oapi-codegen/tree/v2.6.0/pkg/codegen/templates), лицензия Apache 2.0 приложена в `LICENSE`.
Изменены только комментарии: они переведены на русский. Стандартная метка `Code generated ... DO NOT EDIT` сохранена для инструментов Go.

Шаблоны подключены в `../oapi-codegen.yaml`. Команда `make generate` использует их автоматически; `api.gen.go` вручную не редактируется.
При обновлении версии генератора нужно сверить шаблоны с исходными и повторить `make check`.
