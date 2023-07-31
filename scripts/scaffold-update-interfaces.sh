#!/usr/bin/env bash
set -e

# UPDATE INTERFACE FILE
#
# The interface file contains all the API functions we expect to use from the
# go-fastly SDK. When adding a new command, we want to update this file to
# reflect any new API functions we're intending to use.
#
# The logic in this file is more complex than the other scaffolding scripts
# because we're manipulating an existing file that isn't code-generated.
#
# I use Vim to handle the processing because it's easier for me (@integralist)
# to write the otherwise complex logic, compared to trying to use bash or some
# other tool such as Awk.
#
# STEPS:
# - We locate the Interface type.
# - Copy the last set of interface methods.
# - Capture line number for start of copied methods (to use in substitution).
# - Rename the API (three separate places per line).
#
# NOTE:
# Any backslash in the substitution commands (e.g. \v) need to be double escaped.
#   - Once because the backslash is inside the :exe command's expected string.
#   - And then again because of the parent HEREDOC container.
#
# CAVEATS:
# This isn't a perfect process. Its successfulness is based on whether the last
# set of commands align with our expectations. It will still produce ~95%
# expected output, but if there's an extra API function (e.g. BatchModify) then
# that line won't have the relevant API name replaced as we only look for the
# common CRUD methods (Create, Delete, Get, List, Update).
#
vim -E -s pkg/api/interface.go <<-EOF
	:g/type Interface interface/norm $%k
	:norm V{yP]mk
	:norm {
	:call setreg('a', line('.'))
	:norm ]mk
	:exe getreg("a")","line(".")"s/\\\v(Create|Delete|Get|List|Update)[^(]+/\\\1${CLI_API}/"
	:exe getreg("a")","line(".")"s/\\\v(fastly\\\.)(Create|Delete|Get|List|Update)[^)]+(Input)/\\\1\\\2${CLI_API}\\\3/"
	:exe getreg("a")","line(".")"s/\\\v\\\((\\\[\\\])?\\\*(fastly\\\.)[^,]+/(\\\1*\\\2${CLI_API}/"
	:exe getreg("a")","line(".")"s/\\\v(List${CLI_API})/\\\1s/g"
	:update
	:quit
EOF

# The following is essentially the same as above, but we tweak the first :exe
# substitution a bit to fit the format of the mock interface file.
#
vim -E -s pkg/mock/api.go <<-EOF
	:g/type API struct/norm $%k
	:norm V{yP]mk
	:norm {
	:call setreg('a', line('.'))
	:norm ]mk
	:exe getreg("a")","line(".")"s/\\\v(Create|Delete|Get|List|Update)[^(]+/\\\1${CLI_API}Fn func/"
	:exe getreg("a")","line(".")"s/\\\v(fastly\\\.)(Create|Delete|Get|List|Update)[^)]+(Input)/\\\1\\\2${CLI_API}\\\3/"
	:exe getreg("a")","line(".")"s/\\\v\\\((\\\[\\\])?\\\*(fastly\\\.)[^,]+/(\\\1*\\\2${CLI_API}/"
	:exe getreg("a")","line(".")"s/\\\v(List${CLI_API})/\\\1s/g"
	:update
	:quit
EOF

# Additionally, we have to create mock implementations of the CRUD functions,
# so we have to copy an existing function and then do similar substitutions.
#
functions=("Create" "Delete" "Get" "List" "Update")
for fn in "${functions[@]}"; do
	vim -E -s pkg/mock/api.go <<-EOF
		:$
		:norm V{yPG
		:norm {
		:call setreg('a', line('.'))
		:$
		:exe getreg("a")","line(".")"s/\\\v(return m\\\.)(Create|Delete|Get|List|Update)[^(]+/\\\1${fn}${CLI_API}Fn/"
		:exe getreg("a")","line(".")"s/\\\v\\\) (Create|Delete|Get|List|Update)[^(]+/) ${fn}${CLI_API}/"
		:exe getreg("a")","line(".")"s/\\\v(fastly\\\.)(Create|Delete|Get|List|Update)[^)]+(Input)/\\\1${fn}${CLI_API}\\\3/"
		:exe getreg("a")","line(".")"s/\\\v\\\((\\\*fastly\\\.)[^,]+/(\\\1${CLI_API}/"
		:exe getreg("a")","line(".")"s/\\\v^(\\\/\\\/) (Create|Delete|Get|List|Update)(\\\w+)( implements)/\\\1 ${fn}${CLI_API}\\\4/"
		:update
		:quit
	EOF

	# List needs a plural for its name.
	# We can't combine this substitution with the above because of the potential
	# ordering of commands generated (i.e. it could cause another method to be
	# incorrectly updated).
	vim -E -s pkg/mock/api.go <<-EOF
		:$
		:norm {{
		:,+4s/\\v(List${CLI_API})/\\1s/ge
		:update
		:quit
	EOF
done


# UPDATE RUN FILE
#
# The run file contains all the CLI commands we expect to expose to users.
# We want to update this file to reflect any new commands we've added.
#
# STEPS:
# - We locate an existing command we want to copy.
# - Copy the command instantiations.
# - Rename the package name.
# - Yank the new commands to the vim register.
# - Insert the new commands into the list that will be parsed by cmd.Select()
#
# The command we copy depends on whether we're creating a top-level command or
# a category command. If the former we copy the 'backend' command set, if the
# latter we'll copy the 'vcl' command set as it defines the category as a root
# command and passes that to the nested root command.
#
# NOTE:
# Any backslash in the substitution commands need to be escaped because of the
# parent HEREDOC container.
#
# Although it looks like the list of commands in run.go is sorted, they are
# actually manually ordered alphabetically and that's because each commands
# 'root' command needs to be at the top, and sorting the list would cause that
# to break. So it's important you don't attempt to sort the list. The purpose of
# this automation script is to save some manual key strokes. You'll have to
# manually sort the newly created lines yourself.
#
if [[ -z "${CLI_CATEGORY}" ]]; then
vim -E -s pkg/app/commands.go <<-EOF
  :g/backendCmdRoot :=/norm 0
  :norm V5jyP
  :,+5s/backend/${CLI_PACKAGE}/g
  :norm V5k"ay
  :g/return \\[]cmd.Command/norm 0
  :norm "ap
  :,+5s/\\v :=.+/,/
  :norm V5k>
  :update
  :quit
EOF
else
vim -E -s pkg/app/commands.go <<-EOF
  :g/vclCmdRoot :=/norm 0
  :norm V6jyP
  :,+6s/vcl/${CLI_CATEGORY}/g
  :-6
  :,+6s/custom\\./${CLI_PACKAGE}./g
  :-6
  :,+6s/Custom/\\u${CLI_PACKAGE}/g
  :-6
  :norm V6j"ay
  :g/return \\[]cmd.Command/norm 0
  :norm "ap
  :,+6s/\\v :=.+/,/
  :norm V6k>
  :update
  :quit
EOF
fi
