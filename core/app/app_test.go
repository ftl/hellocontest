package app

/*
TEST LIST:

startup
- should start with a default filename
- should start with an empty log
- should update the log view
- should connect the entry view to the new log

new
- should ask for a filename
- should overwrite, if the file already exists
- should update the log view
- should connect the entry view to the new log

open
- should ask for a filename
- should load the log data from file
- should update the log view
- should append new QSOs to the selected file
- should connect the entry view to the loaded log

save as
- should ask for a filename
- should clear the new file
- should overwrite, if the file already exists
- should write all existing QSOs from the log to the file
- should append new QSOs to the selected file

*/
