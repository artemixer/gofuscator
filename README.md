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

<img width="875" alt="Screenshot 2024-02-07 at 23 56 30" src="https://github.com/artemixer/gofuscator/assets/109953672/b961388f-7bfc-44c2-bed9-02fd9adc0615">

<br/>
<br/>

<img width="876" alt="Screenshot 2024-02-07 at 23 57 23 1" src="https://github.com/artemixer/gofuscator/assets/109953672/375e08c6-087a-4cd9-ade4-b3e53fc249fc">

## Functionality
Currently gofuscator is able to process **strings, integers, floats, bools, imports and function/variable names**

Function and variable names, as well as imports, are changed to a random string consisting of the ASCII ```a``` and the cyrilic ```а```, which end up looking visually identical: 
<br/>```str1``` -> ```аaааааaaaaаaaaaaааaa```
<br/>```rand.Read``` -> ```аaааааaaaaаaaaaaааaa.Read```

Bools are changed to a random lesser or greater statement: 
<br/>```false``` -> ```(948 >= 6995)```

Strings are decrypted from a base64 sequence of bytes : 
<br/>```"test"``` -> ```aesDecrypt((string(49) + string(78) + string(57) + ...)```

And all of the above methods are reinforced by the way integers and floats are obfuscated, which is the main feature of this tool.
Integers and floats are converted into a random sequence of mathematical operations, such as ```math.Sqrt```, ```math.Tan``` and others.
The corresponding math functions are called using ```reflect``` to avoid optimisations at compile-time. And finally, all relevant math functions
are cast through a randomly generated function array. This is the result: 
<br/><br/>```-3``` -> ```-(int(aаaааaааaааaaaaааааa.Round((7.809872820273727 * аaаааaaaааaaaaaаaааа.ValueOf(aаaааaааaааaaaaааааa.Pow).Call([]аaаааaaaааaaaaaаaааа.Value{аaаааaaaааaaaaaаaааа.ValueOf(float64(2)), аaаааaaaааaaaaaаaааа.ValueOf(float64(1.8239638712608488))})[0].Interface().(float64)) / аааaаааaааааaаaaaаaa[(int(aаaааaааaааaaaaааааa.Round((аaаааaaaааaaaaaаaааа.ValueOf(aаaааaааaааaaaaааааa.Hypot).Call([]аaаааaaaааaaaaaаaааа.Value{аaаааaaaааaaaaaаaааа.ValueOf(float64(9.872055932101784)), аaаааaaaааaaaaaаaааа.ValueOf(float64(-3.0755290039063317) )})[0].Interface().(float64))*aаaааaааaааaaaaааааa.Cbrt(0.11306904485486288))))](1.4627241154235249))))```

This processing also applies to integers generated at all previous steps.


## Notes
As ```const``` types cannot have values set by functions, they are converted to ```var``` upon processing.


<br/>
<br/>
<b>Feel free to open pull requests and issues, I will do my best to resolve or at least answer all of them</b>
