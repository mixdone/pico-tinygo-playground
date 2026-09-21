# pico-tinygo-playground

Репозиторий для баловства с микроконтроллерами. Эксперименты и
мелкие поделки. 

## Дирректории


| Папка | Что это |
|---|---|
| `display/` | 4-разрядный 7-сегментный дисплей через 74HC595 + 4 кнопки |

## С чем имеем дело

- Raspberry Pi Pico (RP2040)
- TinyGo
- Сдвиговый регистр 74HC595, 7-сегментный дисплей, кнопки, 
  светодиоды, провода

## Сборка и прошивка

Нужен [TinyGo](https://tinygo.org/getting-started/install/) 

### Сборка

```bash
PATH=$HOME/sdk/go1.22.0/bin:$PATH tinygo build -target=pico -o smth.uf2 main.go
```

### Прошивка
``` bash
cp  smth.uf2 /run/media/$USER/RPI-RP2/
```
