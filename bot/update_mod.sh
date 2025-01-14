#!/bin/bash

# Проверка наличия необходимых аргументов
if [ "$#" -ne 2 ]; then
    echo "Использование: $0 <старый_путь> <новый_путь>"
    exit 1
fi

OLD_PATH=$1
NEW_PATH=$2

# Изменение пути в go.mod
if [ -f "go.mod" ]; then
    go mod edit -module $NEW_PATH
else
    echo "Файл go.mod не найден."
    exit 1
fi

# Обновление импортов во всех .go файлах рекурсивно
echo "Обновление импортов..."
find . -type f -name '*.go' -exec sed -i "s|$OLD_PATH|$NEW_PATH|g" {} +

# Запуск go mod tidy для очистки зависимостей
go mod tidy

echo "Путь модуля изменен с '$OLD_PATH' на '$NEW_PATH' и все импорты обновлены."
