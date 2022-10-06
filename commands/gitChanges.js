const gitChanges = async () => {
    const chalk = require('chalk');
	const { simpleGit, CleanOptions } = require('simple-git');
    const gitStatus = await simpleGit().status();
    const gitSummary = await simpleGit().diffSummary();

    if (gitStatus.staged.length + gitSummary.files.length > 0) {
        console.log(chalk.whiteBright('Changes to be committed:'));
        gitStatus.staged.forEach((f) => console.log(`    ${chalk.green(f)}`));
        console.log(chalk.whiteBright('Changes not staged for commit:'));
        gitSummary.files.forEach((f) => console.log(`    ${chalk.red(f.file)}`));
    } else {
        console.log(chalk.blue.bold('No changes staged or unstaged'));
    }
    return null;
};

module.exports = {
    gitChanges,
};

