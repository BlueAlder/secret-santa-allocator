# Secret Santa Allocator

A stupidly over engineered secret santa allocator.

Simple binary to create a secret santa allocation of santas to santees.
Loads two text files and creates a derangement of names to names with a password
so users can input the password and find who their match is.

### Usage

Default usage is providing a config file which defines the names and constraints of the allocation

`secret-santa --configFile config.yaml`

The config file is a yaml file with these fields 

```yaml
# must provide either `file` or `data` field. If both then they are unioned and deduped
names: 
  file: "names.txt" # name of file that contains a list of names
  data: ["sam", "tom", "jim", "grace", "bill"] # array of names that is unioned with the above file

# same as above but for passwords
passwords: 
  file: "passwords.txt"
  data: ["password1", "password2", "password3"]

canAllocateSelf: false # can you be allocated yourself
timeout: 5s # timeout of finding a suitable allocation

# rules per name
rules: 
  - name: "sam" 
    cannotGet:  # will not be allocated
      - "matt"
      - "grace"
  - name: "grace"
    cannotGet:
      - "sam"
      - "olivia c."
```