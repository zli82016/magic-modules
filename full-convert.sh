#!/usr/bin/env bash
############ INPUT ############

betaproviderdirectory=$GOPATH/src/github.com/hashicorp/terraform-provider-google-beta
postswitchovercommit=1056fad8d998ec0647dc2e6123d001e0750e06d8

prnumber=11756
prbranchname=fix-etd-failing-tests-issue-19086

files='
mmv1/third_party/terraform/services/securitycentermanagement/resource_scc_management_organization_event_threat_detection_custom_module_test.go.erb
'

############ END INPUT ############

filelist=$([[ $files =~ [[:space:]]*([^[:space:]]|[^[:space:]].*[^[:space:]])[[:space:]]* ]]; echo "${BASH_REMATCH[1]}")
filelist="${filelist//$'\n'/,}"

echo "Provider directory:            ${betaproviderdirectory}"
echo "Post-switchover commit:        ${postswitchovercommit}"
echo "PR number:                     ${prnumber}"
echo "PR branch name:                ${prbranchname}"
echo "Comma-separated file string:   ${filelist}"

echo "Proceed?"
select yn in "Yes" "No"; do
    case $yn in
        Yes ) break;;
        No ) exit;;
    esac
done

backupbranch="${prbranchname}-backup"

git fetch upstream pull/$prnumber/head:$backupbranch
git checkout $backupbranch
git push --set-upstream origin $backupbranch

git fetch upstream 
git rebase upstream/rewrite-script-updates

sh scripts/convert-go.sh $betaproviderdirectory $filelist

sh scripts/cherry-pick.sh $postswitchovercommit

