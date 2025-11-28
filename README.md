# Инкремент 17. Улучшение производительности по памяти при помощи анализа профилей pprof 

В качестве примера рассматривается оптимизация конфигурации приложения с in-memory storage.

---

1. Для уменьшения аллокаций памяти в `repository.(*URLMemoryStorage).Set` задал подходящий размер слайса `records`, 
в котором хранятся записи хранилища.

2. Для хранилища пользователей также задал подходящий размер карты `users` при инициализации, 
чтобы не происходила постоянная аллокация.

3. В AuthService поменял тип для поля `secret` (`string` -> `[]byte`), 
чтобы избежать аллокации при каждой валидации или создании токена авторизации.

4. Оптимизировал склеивание короткой ссылки (вместо `url.JoinPath` используется кастомный сервис `URLBuilder`), 
количество аллокаций уменьшилось с 7 до 0 (судя по бенчмаркам). 


## Результат сравнения

```shell
 % go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof 
File: shortener
Type: inuse_space
Time: 2025-11-29 01:10:04 MSK
Duration: 150s, Total samples = 35262.66kB 
Showing nodes accounting for -12839.45kB, 36.41% of 35262.66kB total
Dropped 14 nodes (cum <= 176.31kB)
      flat  flat%   sum%        cum   cum%
-5766.29kB 16.35% 16.35% -5766.29kB 16.35%  github.com/alex-storchak/shortener/internal/repository.(*MemoryUserStorage).Set
-5120.31kB 14.52% 30.87% -5536.61kB 15.70%  github.com/alex-storchak/shortener/internal/service.(*Shortener).Shorten
-1024.05kB  2.90% 33.78% -1024.05kB  2.90%  encoding/json.(*decodeState).literalStore
-1024.05kB  2.90% 36.68% -1024.05kB  2.90%  github.com/google/uuid.UUID.String (inline)
 -928.34kB  2.63% 39.31%  -416.29kB  1.18%  github.com/alex-storchak/shortener/internal/repository.(*MemoryURLStorage).Set
    -513kB  1.45% 40.77%     -513kB  1.45%  sync.(*Pool).pinSlow
  512.50kB  1.45% 39.32%   512.50kB  1.45%  go.uber.org/zap/internal/bufferpool.init.NewPool.func1
  512.17kB  1.45% 37.86%   512.17kB  1.45%  net/textproto.MIMEHeader.Add (inline)
 -512.16kB  1.45% 39.32%  -512.16kB  1.45%  net/http.(*Request).WithContext (inline)
  512.06kB  1.45% 37.86%   512.06kB  1.45%  internal/profile.(*Profile).postDecode
 -512.05kB  1.45% 39.31%  -512.05kB  1.45%  internal/profile.init.func5
  512.05kB  1.45% 37.86%   512.05kB  1.45%  internal/sync.runtime_SemacquireMutex
  512.01kB  1.45% 36.41%   512.01kB  1.45%  github.com/teris-io/shortid.(*Abc).Encode
```