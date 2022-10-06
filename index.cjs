#!/usr/bin/env node
const { program } = require('commander');
const { gitChanges } = require('./commands/gitChanges');

program
    .command('git-changes')
    .description('List all analyzable files modified since last git commit')
    .action(gitChanges);

program.parse();
