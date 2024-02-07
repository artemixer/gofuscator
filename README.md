# gofuscator
**gofuscator** is an obfuscator/polymorphic engine for Golang code. 


## Installation
Make sure you have Golang installed by running ```go version```
```
git clone https://github.com/artemixer/gofuscator
cd gofuscator
go build gofuscator.go
```
  
## Usage
```
./gofuscator -i input_file.go -o output_file.go
```
Here is a sample before and after the obfuscation process:

<img width="922" alt="Screenshot 2024-02-07 at 14 13 28" src="https://github.com/artemixer/gofuscator/assets/109953672/8e36ebbc-fc2c-4211-8835-96aca05e7696">

<br/>
<br/>

<img width="923" alt="Screenshot 2024-02-07 at 14 13 44" src="https://github.com/artemixer/gofuscator/assets/109953672/dafcc981-47a7-450c-8dce-1325a83b15a6">


## Notes
As ```const``` types cannot have values set by functions, they are converted to ```var``` upon processing.


<br/>
<br/>
<b>Feel free to open pull requests and issues, I will do my best to resolve or at least answer all of them</b>
