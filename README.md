# Бот для клубов

## Todo
- подключить линтер
- накидать структуру Golang проетка для тг-бота
- схему бд и бота сюда залить 

[Схема бота](https://miro.com/welcomeonboard/NnB1K3UwMHg4TEIrWUFkRkwyT2ZlRDI5dW5WSm8xeVltNlNIWFp0b055a3lVVm5SY2t5TENEWFZsNTVMVUxxalp4V256dlA0SnBRbEQxQ2YxcmZVSHI4UXQ4TkQ1d3hQMjlqWmQwZi9DTHBNeWZJaVdvU0pJSGErNytVZDJ0amkhZQ==?share_link_id=868943904438)


Схема БД
```mermaid
erDiagram
    USERS {
        uuid id PK "Уникальный ID пользователя"
        bigint telegram_id not null "Telegram ID пользователя"
        string role not null "Роль: student, grant_user, external_user"
        string email "Email для студентов"
        string fio not null "ФИО пользователя"
        bool is_club_owner default false
        timestamp created_at default current_timestamp "Дата создания записи"
        timestamp updated_at
        boolean is_banned default true "Активность пользователя"
    }

    CLUBS {
        uuid id PK "Уникальный ID клуба"
        string name not null "Название клуба"
        string description not null  "Описание клуба"
        timestamp created_at default current_timestamp "Дата создания"
        timestamp updated_at
        timestamp deleted_at
    }

    CLUB_OWNERS {
        uuid user_id PK "ID пользователя (связь с USERS)"
        uuid club_id PK "ID клуба (связь с CLUBS)"
        timestamp created_at default current_timestamp "Дата назначения роли"
    }

    EVENTS {
        uuid id PK "Уникальный ID мероприятия"
        uuid club_id FK "ID клуба (связь с CLUBS)"
        string name not null "Название мероприятия"
        string description not null "Описание мероприятия"
        string after_registration_text
        string location not null "Место проведения"
        timestamp start_time not null "Время начала"
        timestamp end_time "Время окончания"
        timestamp registration_end "Время окончания регистрации"
        int max_participants "Максимальное количество участников"
        int expected_participants not null "Ожидаемое число участников"
        timestamp created_at default current_timestamp "Дата создания"
        timestamp updated_at
        timestamp deleted_at
    }

    EVENT_PARTICIPANTS {
        uuid event_id PK "ID мероприятия (связь с EVENTS)"
        uuid user_id PK "ID пользователя (связь с USERS)"
        bool user_qr "Пришёл ли пользователь с своего qr"
        bool event_qr "Пришёл ли пользователь с qr мероприятия"
        timestamp registered_at default current_timestamp "Дата регистрации"
    }

    %% Relationships
    USERS ||--o{ CLUB_OWNERS : "can own"
    CLUBS ||--o{ CLUB_OWNERS : "is owned by"
    CLUBS ||--o{ EVENTS : "organizes"
    EVENTS ||--o{ EVENT_PARTICIPANTS : "has participants"
    USERS ||--o{ EVENT_PARTICIPANTS : "participates in"
```
