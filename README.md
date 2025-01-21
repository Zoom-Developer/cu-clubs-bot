# Бот для клубов

[Схема бота](https://miro.com/welcomeonboard/NnB1K3UwMHg4TEIrWUFkRkwyT2ZlRDI5dW5WSm8xeVltNlNIWFp0b055a3lVVm5SY2t5TENEWFZsNTVMVUxxalp4V256dlA0SnBRbEQxQ2YxcmZVSHI4UXQ4TkQ1d3hQMjlqWmQwZi9DTHBNeWZJaVdvU0pJSGErNytVZDJ0amkhZQ==?share_link_id=868943904438)


Схема БД
```mermaid
erDiagram
    USERS {
        uuid id PK "Уникальный ID пользователя"
        bigint telegram_id "Telegram ID пользователя"
        string role "Роль: student, grant_user, external_user"
        string email "Email для студентов"
        string fio "ФИО пользователя"
        timestamp created_at "Дата создания"
        timestamp updated_at "Дата обновления"
        boolean is_banned "Активность пользователя"
    }

    CLUBS {
        uuid id PK "Уникальный ID клуба"
        string name "Название клуба"
        string description "Описание клуба"
        timestamp created_at "Дата создания"
        timestamp updated_at "Дата обновления"
        timestamp deleted_at "Дата удаления"
    }

    CLUB_OWNERS {
        uuid user_id PK "ID пользователя (связь с USERS)"
        uuid club_id PK "ID клуба (связь с CLUBS)"
        timestamp created_at "Дата назначения роли"
    }

    EVENTS {
        uuid id PK "Уникальный ID мероприятия"
        uuid club_id FK "ID клуба (связь с CLUBS)"
        string name "Название мероприятия"
        string description "Описание мероприятия"
        string after_registration_text "Текст, показывающийся пользователю после регистрации на мероприятие"
        string location "Место проведения"
        timestamp start_time "Время начала"
        timestamp end_time "Время окончания"
        timestamp registration_end "Время окончания регистрации"
        int max_participants "Максимальное количество участников"
        int expected_participants "Ожидаемое число участников"
        timestamp created_at "Дата создания"
        timestamp updated_at "Дата обновления"
        timestamp deleted_at "Дата удаления"
    }

    EVENT_PARTICIPANTS {
        uuid event_id PK "ID мероприятия (связь с EVENTS)"
        uuid user_id PK "ID пользователя (связь с USERS)"
        bool user_qr "Пришёл ли пользователь с своего qr"
        bool event_qr "Пришёл ли пользователь с qr мероприятия"
        timestamp registered_at "Дата регистрации"
    }

    %% Relationships
    USERS ||--o{ CLUB_OWNERS : "can own"
    CLUBS ||--o{ CLUB_OWNERS : "is owned by"
    CLUBS ||--o{ EVENTS : "organizes"
    EVENTS ||--o{ EVENT_PARTICIPANTS : "has participants"
    USERS ||--o{ EVENT_PARTICIPANTS : "participates in"
```
