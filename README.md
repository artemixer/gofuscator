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

<img width="922" alt="Screenshot 2024-02-06 at 22 52 28" src="https://github.com/artemixer/gofuscator/assets/109953672/3e45c7a4-fb37-42c1-9bf2-433f1af9d26c">
<br/>
<br/>

<img width="921" alt="Screenshot 2024-02-06 at 22 52 40" src="https://github.com/artemixer/gofuscator/assets/109953672/8458afc2-fc29-45de-b590-7b3a665e3f96">

## Notes
As ```const``` types cannot have values set by functions, they are converted to ```var``` upon processing.


<br/>
<br/>
<b>Feel free to open pull requests and issues, I will do my best to resolve or at least answer all of them</b>
