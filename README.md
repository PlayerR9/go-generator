# go-generator
A Go package used for generating Go codes.


### Usage

***Initialization***

First and foremost, you have to initialize the template. You can do like this:
```go
import (
   "text/template"
)

type GenData struct {
   // Initialize here the various fields that will be used in the template.
}

var (
   my_template *template.Template
   my_generator *ggen.CodeGenerator
)

func init() {
   my_template = template.Must(template.New("").Parse(templ))

   tmp, err := ggen.NewCodeGenerator[*GenData](my_template)
   if err != nil {
      // Handle the error. (e.g. print the error message)
   }

   my_generator = tmp

   // Set the various do functions if needed.
   // my_generator.SetDoFunc(func(g *GenData) error {
   //    // Do something
   //
   //    return nil
   // }
}

const templ = "my template"
```

Or you can do like this:
```go
import (
   "text/template"
)

type GenData struct {
   // Initialize here the various fields that will be used in the template.
}

var (
   my_generator *ggen.CodeGenerator
)

func init() {
   tmp, err := ggen.NewCodeGeneratorFromTemplate[*GenData]("", templ)
   if err != nil {
      // Handle the error. (e.g. print the error message)
   }

   my_generator = tmp

   // Set the various do functions if needed.
   // my_generator.SetDoFunc(func(g *GenData) error {
   //    // Do something
   //
   //    return nil
   // }
}

const templ = "my template"
```

The latter is much easier.


***Setup***

Then, you have to initialize the various flags. The only mandatory flag is `ggen.SetOutputFlag()` as it specifies the location of the output file.

After that, you have to declare a struct that will be used in the template. For this example, we will use the struct `GenData`. This struct must implement the `SetPackageName(pkg_name string) ggen.Generater` method as it will be used to set the package name for the generated code.


***Main Function***

The main function is composed of the following steps:
1. **Flag Parsing:** Call the `ggen.ParseFlags()` method to parse the command-line flags and any flag set by the user.
2. **Validation:** Do some validation or sanity checks.
3. **Fixing:** Call the `ggen.FixOutputLoc()` method to fix the output location. This will validate and fix the output path if necessary.
4. **Generation:** Call the `ggen.Generate()` method to generate the code. This method requires the following parameters:
   - `output_loc`: The location of the output file.
   - `data`: The data that will be used in the template. Ideally, this should only be an empty struct as any other field should be initialized in the functions. However, for complex checks or other reasons, they can be initialized in the "validation" step.
   - `template`: The template that will be used to generate the code.
   - `doFunc`: Here, you can specify any function that will be called right before generating the code. It is suggested to initialize all the struct fields in this portion.


Here's an example of a simple usage:
```go
func main() {
   // 1. Call ParseFlags() method to parse the command-line flags.
	err := ggen.ParseFlags()
	if err != nil {
		// Handle the error. (e.g. print the error message)
	}

	// 2. Do some validation.

   // 3. Generate the code.
   data, err := my_generator.Generate(
      "foo", "_template.go",
      &GenData{
         // Initialize here the initial values of the struct that will be used in the template.
      },
   )
	if err != nil {
		// Handle the error.
	}

   // 4. Print the generated code to either stdout or a file.
}
```


### Additional Notes

***Logging***

A common way to handle errors is to log them with the `log.Fatalf()` function. Because of that, I provided the `InitLogger()` function that initializes a common logger.

```go
my_logger := ggen.InitLogger(os.Stdout, "my_logger")
```

This code will create a new logger that will print the file and line number where the error occurred. This is equivalent to do as follows:
```go
my_logger := *log.New(os.Stdout, "[my_logger]: ", log.Lshortfile)
```


***Naming Validation***

Usually, generation requires a name of the type that is generated. To simplify this process, I provided the `IsValidName()` function that checks if the name is valid. Here's an example:
```go
err = ggen.IsValidName("foo", []string{"bar"}, ggen.Exported)
if err != nil {
	// Handle the error. (e.g. print the error message)
}
```

Here's a breakdown:
- `"foo"`: This parameter is the name given by the user and, often, passed through a flag. For instance, if we wanted to generate a function named `foo`, we would pass `"foo"` as the parameter.
- `[]string{"bar"}`: This parameter is a list of all names that cannot be used as a valid name. For instance, if the user specified a name that would conflict with the name of an existing function, we would pass the name of that function as a member of the list. *IMPORTANT: Go keywords are already excluded and so, it is not necessary to include them in the list. Thus, names such as "var", "chan", and so on are never allowed.*
- `ggen.Exported`: This parameter checks the casing of the name. More specifically:
   - `ggen.Exported` will check if the name starts with an uppercase letter.
   - `ggen.NotExported` will check if the name starts with a lowercase letter.
   - `ggen.Either` will not check if the name starts with an uppercase letter or a lowercase letter.

Therefore, if you need a lowercase name that is specified by the user and not "foo", you would specify the parameter as follows:
```go
var (
   type_name *string = flag.String("type", "", "The type of the linked stack.")
)

err = ggen.IsValidName(type_name, []string{"foo"}, ggen.NotExported)
if err != nil {
   // Handle the error. (e.g. print the error message)
}
```


***Type Signatures***

For handling type signatures, I provided the `MakeTypeSig()` function. Here's an example:
```go
sig, err := ggen.MakeTypeSig("foo", "bar")
if err != nil {
   // Handle the error. (e.g. print the error message)
}
```

This function will create a type signature "foobar" because the first parameter is the name of the type and the second parameter is the suffix.

One particular thing of this function is that it also handles generics. In fact, if the flag `GenericsSigFlag` is set, then the function will add the generics signature to the type signature. As such, it is useful for handling methods that have generics.


For example:
```go
// Assuming that the GenericsSigFlag is set as: "foo/T,bar/C"

sig, err := ggen.MakeTypeSig("foo", "bar")
if err != nil {
   // Handle the error. (e.g. print the error message)
}

// The output will be: "foobar[T, C]"
```

Likewise, to have the full type signature, we can call the `String()` method of the `GenericsSigFlag`:
```go
// Assuming that the GenericsSigFlag is set as: "foo/T,bar/C"

type_name := "foobar" + GenericsSigFlag.String()

// The output will be: "foobar[T any, C any]"
```

Of course, if the `GenericsSigFlag` is set but no generics are specified, then the function will return an empty string.
