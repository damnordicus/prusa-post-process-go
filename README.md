# prusa-post-process-go
post process program for filament inventory manager in golang

## Possible values
```
; filament used [g] = 94.13
; printer_model = MK3S
printer_model=XL5
filament used [g]=73.27, 0.00, 0.00, 0.00, 0.00
```
Not: 
```
; output_filename_format = {input_filename_base}_{layer_height}mm_{printing_filament_types}_{printer_model}_{print_time}.gcode
; total filament used [g] = 94.13
; total filament used for wipe tower [g] = 0.00
total filament used for wipe tower [g]=0.00
```
