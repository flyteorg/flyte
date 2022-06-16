## Customize NavBar component

From this point forward you can modify your FlyteConsole navigatio bar by:

-   using your company colors
-   providing entrypoint navigation to sites, or places inside flyteconsole.

To use it you will need to define `FLYTE_NAVIGATION` environment variable during the build.

If you are building locally add next or similar export to your `.zshrc` (or equivalent) file:

```bash
export FLYTE_NAVIGATION='{"color":"white","background":"black","items":[{"title":"Hosted","url":"https://hosted.cloud-staging.union.ai/dashboard"}, {"title":"Dashboard","url":"/projects/flytesnacks/executions?domain=development&duration=all"},{"title":"Execution", "url":"/projects/flytesnacks/domains/development/executions/awf2lx4g58htr8svwb7x?duration=all"}]}'
```

If you are building a docker image - modify [Makefile](./Makefile) `build_prod` step to include FLYTE_NAVIGATION setup.

### The structure of FLYTE_NAVIGATION

Essentially FLYTE_NAVIGATION is a JSON object

```
{
    color:"white",       // default NavBar text color
    background:"black",  // default NavBar background color
    items:[
        {title:"Remote", url:"https://remote.site/"},
        {title:"Dashboard", url:"/projects/flytesnacks/executions?domain=development&duration=all"},
        {title:"Execution", url:"/projects/flytesnacks/domains/development/executions/awf2lx4g58htr8svwb7x?duration=all"}
    ]
}
```

If at least one item in `items` array is present the dropdown will appear in NavBar main view.
It will contain at least two items:

-   default "Console" item which navigates to ${BASE_URL}
-   all items you have provided

Feel free to play around with the views in Storybook:

<img width="874" alt="Screen Shot 2022-06-15 at 2 01 29 PM" src="https://user-images.githubusercontent.com/55718143/173962811-a3603d6c-3fe4-4cab-b57a-4d4806c88cfc.png">


#### Note

Please let us know in [Slack #flyte-console](https://flyte-org.slack.com/archives/CTJJLM8BY) channel if you found bugs or need more support than.
