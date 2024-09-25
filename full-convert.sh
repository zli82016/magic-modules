#!/usr/bin/env bash
############ INPUT ############
magic_modules_directory=/usr/local/google/home/thomasrodgers/go/src/github.com/trodge/magic-modules/tmp-testing
post_switchover_commit=a520431d52493a87387c167f28cadac63d2e2e0e
beta_provider_directory=$GOPATH/src/github.com/hashicorp/terraform-provider-google-beta

pr_number=11639
pr_branch_name=main
legacy_branch=upstream/legacy-ruby

############ END INPUT ############

echo "Provider directory:            ${beta_provider_directory}"
echo "Post-switchover commit:        ${post_switchover_commit}"
echo "PR number:                     ${pr_number}"
echo "PR branch name:                ${pr_branch_name}"
echo "Legacy branch:                 ${legacy_branch}"


echo "Proceed?"
select yn in "Yes" "No"; do
    case $yn in
        Yes ) break;;
        No ) exit;;
    esac
done

cd $magic_modules_directory

backup_branch="${pr_branch_name}-backup"

git fetch upstream pull/$pr_number/head:$backup_branch
git checkout $backup_branch
git push --set-upstream origin $backup_branch

backup_branch="$(git rev-parse --symbolic-full-name --abbrev-ref HEAD)"
if [[ $backup_branch != *"-backup"* ]]; then
  echo "\"-backup\" not detected in the branch name \"${backup_branch}\""
  echo "Do you want to still continue with the default branch name \"go-rewrite-convert\"?"
  select yn in "Yes" "No"; do
    case $yn in
        Yes ) backup_branch="go-rewrite-convert"; break;;
        No ) exit;;
    esac
  done
fi
new_branch=${backup_branch%"-backup"}
echo "will use branch \"${new_branch}\""

git fetch upstream 
git rebase $legacy_branch

files=$(git diff --name-only $legacy_branch)

file_list="${files//$'\n'/,}"

echo "Changed files:"
echo "${files} "
echo "Comma-separated file string:   ${file_list}"

echo "Proceed?"
select yn in "Yes" "No"; do
    case $yn in
        Yes ) break;;
        No ) exit;;
    esac
done

############ file converting section ############

yaml_list=()
erb_list=()
other_list=()
yaml_string=""
erb_string=""

# Read all input files
IFS=',' read -ra file <<< "$file_list"
for i in "${file[@]}"; do
  filename=$(basename -- "$i")
  extension="${filename##*.}"
  if [[ $extension == "yaml" ]]; then
    yaml_list+=($i)
    if [[ $yaml_string == "" ]]; then
        yaml_string=$i
    else
        yaml_string="${yaml_string},${i}"
    fi
  elif [[ $extension == "erb" ]]; then
    erb_list+=($i)
    if [[ $erb_string == "" ]]; then
        erb_string=$i
    else
        erb_string="${erb_string},${i}"
    fi
  else
    other_list+=($i)
  fi
done

pushd mmv1

if [[ $yaml_string != "" ]]; then
  # run yaml conversion with given .yaml files
  bundle exec compiler.rb -e terraform -o $beta_provider_directory -v beta -a --go-yaml-files $yaml_string
  go run . --yaml-temp
  for i in `find . -name "*.temp" -type f`; do
    echo "removing go/ paths in ${i}"
    perl -pi -e 's/go\///g' $i
  done

fi

if [[ $erb_string != "" ]]; then
  # convert .erb files with given .erb files
  go run . --template-temp $erb_string
  go run . --handwritten-temp $erb_string
fi
popd

# add temporary file for all other files that do not need conversion
for i in "${other_list[@]}"
do
    cp "$i" "${i}.temp" 
done


############ cherry-picking section ############

# prepare temp file commit
git add .
git commit -m "temp file commit" 
current_commit="$(git rev-parse HEAD)"
echo "committed all changes to ${current_commit}"

# checkout a new branch from a given post-switchover commit
echo "checking out \"${new_branch}\" at post-switchover commit ${post_switchover_commit}"
git checkout -b $new_branch $post_switchover_commit

# cherry-pick the previous temp file commit to the new post-switchover branch
echo "cherry-picking ${current_commit}"
git cherry-pick $current_commit --no-commit

echo "Merge conflicts resolved?"
select yn in "Yes" "No"; do
    case $yn in
        Yes ) break;;
        No ) exit;;
    esac
done

# overwrite the converted files with the temporary files to produce final diff
echo "cherry-picking ${current_commit}"
files=`git diff --name-only --diff-filter=A --cached`
for file in $files; do
  echo "moving ${file} to ${file%".temp"}"
  mv $file ${file%".temp"}
done

# stage all changes
git add .
git status
