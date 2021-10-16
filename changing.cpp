/*  C++ Program to Convert Days Into Years Weeks and Days  */

#include<iostream>

using namespace std;

int main()
{
    int y,d,w;

    cout<<"Enter No. of days :: ";
    cin>>d;

    y=d/365;
    d=d%365;
    w=d/7;
    d=d%7;

    cout<<"\nNo. of Years: : "<<y<<"\nNo. of Weeks :: "<<w<<"\nNo. of Days :: "<<d<<"\n";

    return 0;
}
