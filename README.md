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

5. Заменил кодирование/декодирование для запросов на `easyjson`.

## Результат сравнения профилей до и после:

```shell
% go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof                
File: shortener
Type: inuse_space
Time: 2025-11-29 13:20:23 MSK
Duration: 150s, Total samples = 27087.48kB 
Showing nodes accounting for -6598.84kB, 24.36% of 27087.48kB total
      flat  flat%   sum%        cum   cum%
-9216.42kB 34.02% 34.02% -9216.42kB 34.02%  encoding/json.(*decodeState).literalStore
 7680.35kB 28.35%  5.67%  7680.35kB 28.35%  github.com/mailru/easyjson/jlexer.(*Lexer).String
-5120.31kB 18.90% 24.57% -5483.70kB 20.24%  github.com/alex-storchak/shortener/internal/service.(*Shortener).Shorten
   -3678kB 13.58% 38.15%    -3678kB 13.58%  github.com/alex-storchak/shortener/internal/repository.(*MemoryUserStorage).Set
 2560.12kB  9.45% 28.70%  2560.12kB  9.45%  github.com/google/uuid.UUID.String (inline)
-1899.41kB  7.01% 35.71% -1899.41kB  7.01%  github.com/alex-storchak/shortener/internal/repository.(*MemoryURLStorage).Set
 1024.02kB  3.78% 31.93%  1536.02kB  5.67%  github.com/teris-io/shortid.(*Shortid).GenerateInternal
  516.01kB  1.90% 30.03%   516.01kB  1.90%  io.init.func1
    -513kB  1.89% 31.92%     -513kB  1.89%  bufio.NewWriterSize (inline)
    -513kB  1.89% 33.82%     -513kB  1.89%  sync.(*Pool).pinSlow
  512.50kB  1.89% 31.92%   512.50kB  1.89%  go.uber.org/zap/internal/bufferpool.init.NewPool.func1
  512.14kB  1.89% 30.03%   512.14kB  1.89%  strings.(*Builder).grow
  512.12kB  1.89% 28.14%   512.12kB  1.89%  github.com/go-chi/chi/v5.NewRouteContext
  512.05kB  1.89% 26.25%   512.05kB  1.89%  sync.runtime_notifyListWait
  512.01kB  1.89% 24.36%   512.01kB  1.89%  github.com/teris-io/shortid.(*Abc).Encode
```

## Вывод

Общее снижение потребления памяти снизилось на 6.5 MB при общем сэмпле в 27 MB, что составляет 24%. 